package setup

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"time"
)

// SQLExecutor 安装或升级SQL执行器
type SQLExecutor struct {
	dbConfig  *dbs.DBConfig
	logWriter io.Writer
}

func NewSQLExecutor(dbConfig *dbs.DBConfig) *SQLExecutor {
	return &SQLExecutor{
		dbConfig: dbConfig,
	}
}

func NewSQLExecutorFromCmd() (*SQLExecutor, error) {
	// 执行SQL
	var config = &dbs.Config{}
	configData, err := os.ReadFile(Tea.ConfigFile("db.yaml"))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, err
	}
	return NewSQLExecutor(config.DBs[Tea.Env]), nil
}

func (this *SQLExecutor) SetLogWriter(logWriter io.Writer) {
	this.logWriter = logWriter
}

func (this *SQLExecutor) Run(showLog bool) error {
	db, err := dbs.NewInstanceFromConfig(this.dbConfig)
	if err != nil {
		return err
	}

	// prevent default configure loading
	var globalConfig = dbs.GlobalConfig()
	if globalConfig != nil && len(globalConfig.DBs) == 0 {
		globalConfig.DBs = map[string]*dbs.DBConfig{"prod": this.dbConfig}
	}

	defer func() {
		_ = db.Close()
	}()

	var sqlDump = NewSQLDump()
	sqlDump.SetLogWriter(this.logWriter)
	if this.logWriter != nil {
		showLog = true
	}

	var sqlResult = &SQLDumpResult{}
	err = json.Unmarshal(sqlData, sqlResult)
	if err != nil {
		return fmt.Errorf("decode sql data failed: %w", err)
	}

	_, err = sqlDump.Apply(db, sqlResult, showLog)
	if err != nil {
		return err
	}

	// 检查数据
	err = this.checkData(db)
	if err != nil {
		return err
	}

	return nil
}

// 检查数据
func (this *SQLExecutor) checkData(db *dbs.DB) error {
	// 检查管理员平台节点
	err := this.checkAdminNode(db)
	if err != nil {
		return fmt.Errorf("check admin node failed: %w", err)
	}

	// 检查用户平台节点
	err = this.checkUserNode(db)
	if err != nil {
		return fmt.Errorf("check user node failed: %w", err)
	}

	// 检查集群配置
	err = this.checkCluster(db)
	if err != nil {
		return fmt.Errorf("check cluster failed: %w", err)
	}

	// 检查初始化用户
	// 需要放在检查集群后面
	err = this.checkUser(db)
	if err != nil {
		return fmt.Errorf("check user failed: %w", err)
	}

	// 检查IP名单
	err = this.checkIPList(db)
	if err != nil {
		return fmt.Errorf("check ip list failed: %w", err)
	}

	// 检查指标设置
	err = this.checkMetricItems(db)
	if err != nil {
		return fmt.Errorf("check metric items failed: %w", err)
	}

	// 检查自建DNS全局设置
	err = this.checkNS(db)
	if err != nil {
		return fmt.Errorf("check ns failed: %w", err)
	}

	// 更新Agents
	err = this.checkClientAgents(db)
	if err != nil {
		return fmt.Errorf("check client agents failed: %w", err)
	}

	// 更新版本号
	err = this.updateVersion(db, ComposeSQLVersion())
	if err != nil {
		return fmt.Errorf("update version failed: %w", err)
	}

	return nil
}

// 创建初始用户
func (this *SQLExecutor) checkUser(db *dbs.DB) error {
	one, err := db.FindOne("SELECT id FROM edgeUsers LIMIT 1")
	if err != nil {
		return err
	}
	if len(one) > 0 {
		return nil
	}

	// 读取默认集群ID
	// Read default cluster id
	clusterId, err := db.FindCol(0, "SELECT id FROM edgeNodeClusters WHERE state=1 ORDER BY id ASC LIMIT 1")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO edgeUsers (`username`, `password`, `fullname`, `isOn`, `state`, `createdAt`, `clusterId`) VALUES (?, ?, ?, ?, ?, ?, ?)", "USER_"+rands.HexString(10), stringutil.Md5(rands.HexString(32)), "默认用户", 1, 1, time.Now().Unix(), clusterId)
	return err
}

// 检查管理员平台节点
func (this *SQLExecutor) checkAdminNode(db *dbs.DB) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeAPITokens WHERE role='admin'")
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	col, err := stmt.FindCol(0)
	if err != nil {
		return err
	}
	var count = types.Int(col)
	if count > 0 {
		return nil
	}

	var nodeId = rands.HexString(32)
	var secret = rands.String(32)
	_, err = db.Exec("INSERT INTO edgeAPITokens (nodeId, secret, role) VALUES (?, ?, ?)", nodeId, secret, "admin")
	if err != nil {
		return err
	}

	return nil
}

// 检查用户平台节点
func (this *SQLExecutor) checkUserNode(db *dbs.DB) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeAPITokens WHERE role='user'")
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	col, err := stmt.FindCol(0)
	if err != nil {
		return err
	}
	var count = types.Int(col)
	if count > 0 {
		return nil
	}

	var nodeId = rands.HexString(32)
	var secret = rands.String(32)
	_, err = db.Exec("INSERT INTO edgeAPITokens (nodeId, secret, role) VALUES (?, ?, ?)", nodeId, secret, "user")
	if err != nil {
		return err
	}

	return nil
}

// 检查集群配置
func (this *SQLExecutor) checkCluster(db *dbs.DB) error {
	/// 检查是否有集群数据
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeNodeClusters")
	if err != nil {
		return fmt.Errorf("query clusters failed: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	col, err := stmt.FindCol(0)
	if err != nil {
		return fmt.Errorf("query clusters failed: %w", err)
	}
	var count = types.Int(col)
	if count > 0 {
		return nil
	}

	// 创建默认集群
	var uniqueId = rands.HexString(32)
	var secret = rands.String(32)

	var clusterDNSConfig = &dnsconfigs.ClusterDNSConfig{
		NodesAutoSync:    true,
		ServersAutoSync:  true,
		CNAMERecords:     []string{},
		CNAMEAsDomain:    true,
		TTL:              0,
		IncludingLnNodes: true,
	}
	clusterDNSConfigJSON, err := json.Marshal(clusterDNSConfig)
	if err != nil {
		return err
	}

	var defaultDNSName = "g" + rands.HexString(6) + ".cdn"
	{
		var b = make([]byte, 3)
		_, err = rand.Read(b)
		if err == nil {
			defaultDNSName = fmt.Sprintf("g%x.cdn", b)
		}
	}

	_, err = db.Exec("INSERT INTO edgeNodeClusters (name, useAllAPINodes, state, uniqueId, secret, dns, dnsName) VALUES (?, ?, ?, ?, ?, ?, ?)", "默认集群", 1, 1, uniqueId, secret, string(clusterDNSConfigJSON), defaultDNSName)
	if err != nil {
		return err
	}

	// 创建APIToken
	_, err = db.Exec("INSERT INTO edgeAPITokens (nodeId, secret, role, state) VALUES (?, ?, 'cluster', 1)", uniqueId, secret)
	if err != nil {
		return err
	}

	// 默认缓存策略

	models.SharedHTTPCachePolicyDAO = models.NewHTTPCachePolicyDAO()
	models.SharedHTTPCachePolicyDAO.Instance = db
	policyId, err := models.SharedHTTPCachePolicyDAO.CreateDefaultCachePolicy(nil, "默认集群")
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE edgeNodeClusters SET cachePolicyId=?", policyId)
	if err != nil {
		return err
	}

	// 默认WAf策略
	models.SharedHTTPFirewallPolicyDAO = models.NewHTTPFirewallPolicyDAO()
	models.SharedHTTPFirewallPolicyDAO.Instance = db

	models.SharedHTTPFirewallRuleGroupDAO = models.NewHTTPFirewallRuleGroupDAO()
	models.SharedHTTPFirewallRuleGroupDAO.Instance = db

	models.SharedHTTPFirewallRuleSetDAO = models.NewHTTPFirewallRuleSetDAO()
	models.SharedHTTPFirewallRuleSetDAO.Instance = db

	models.SharedHTTPFirewallRuleDAO = models.NewHTTPFirewallRuleDAO()
	models.SharedHTTPFirewallRuleDAO.Instance = db

	models.SharedHTTPWebDAO = models.NewHTTPWebDAO()
	models.SharedHTTPWebDAO.Instance = db

	models.SharedServerDAO = models.NewServerDAO()
	models.SharedServerDAO.Instance = db

	models.SharedNodeClusterDAO = models.NewNodeClusterDAO()
	models.SharedNodeClusterDAO.Instance = db

	policyId, err = models.SharedHTTPFirewallPolicyDAO.CreateDefaultFirewallPolicy(nil, "默认集群")
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE edgeNodeClusters SET httpFirewallPolicyId=?", policyId)
	if err != nil {
		return err
	}

	return nil
}

// 检查IP名单
func (this *SQLExecutor) checkIPList(db *dbs.DB) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeIPLists")
	if err != nil {
		return fmt.Errorf("query ip lists failed: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	col, err := stmt.FindCol(0)
	if err != nil {
		return fmt.Errorf("query ip lists failed: %w", err)
	}
	var count = types.Int(col)
	if count > 0 {
		return nil
	}

	// 创建名单
	_, err = db.Exec("INSERT INTO edgeIPLists(name, type, code, isPublic, isGlobal, createdAt) VALUES (?, ?, ?, ?, ?, ?)", "公共黑名单", "black", "black", 1, 1, time.Now().Unix())
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO edgeIPLists(name, type, code, isPublic, isGlobal, createdAt) VALUES (?, ?, ?, ?, ?, ?)", "公共白名单", "white", "white", 1, 1, time.Now().Unix())
	if err != nil {
		return err
	}

	return nil
}

// 检查统计指标
func (this *SQLExecutor) checkMetricItems(db *dbs.DB) error {
	var createMetricItem = func(code string,
		category string,
		name string,
		keys []string,
		period int,
		periodUnit string,
		value string,
		chartMaps []maps.Map,
	) error {
		// 检查是否已创建
		itemMap, err := db.FindOne("SELECT id FROM edgeMetricItems WHERE code=? LIMIT 1", code)
		if err != nil {
			return err
		}

		if len(itemMap) == 0 {
			keysJSON, err := json.Marshal(keys)
			if err != nil {
				return err
			}
			_, err = db.Exec("INSERT INTO edgeMetricItems (isOn, code, category, name, `keys`, period, periodUnit, value, state, isPublic) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", 1, code, category, name, keysJSON, period, periodUnit, value, 1, 1)
			if err != nil {
				return err
			}

			// 再次查询
			itemMap, err = db.FindOne("SELECT id FROM edgeMetricItems WHERE code=? LIMIT 1", code)
			if err != nil {
				return err
			}
		}

		var itemId = itemMap.GetInt64("id")

		// chart
		for _, chartMap := range chartMaps {
			var chartCode = chartMap.GetString("code")
			one, err := db.FindOne("SELECT id FROM edgeMetricCharts WHERE itemId=? AND code=? LIMIT 1", itemId, chartCode)
			if err != nil {
				return err
			}
			if len(one) == 0 {
				_, err = db.Exec("INSERT INTO edgeMetricCharts (itemId, name, code, type, widthDiv, params, isOn, state) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", itemId, chartMap.GetString("name"), chartCode, chartMap.GetString("type"), chartMap.GetInt("widthDiv"), "{}", 1, 1)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	{
		err := createMetricItem("ip_requests", serverconfigs.MetricItemCategoryHTTP, "独立IP请求数", []string{"${remoteAddr}"}, 1, "day", "${countRequest}", []maps.Map{
			{
				"name":     "独立IP排行",
				"type":     "bar",
				"widthDiv": 0,
				"code":     "ip_requests_bar",
			},
		})
		if err != nil {
			return err
		}
	}

	{
		err := createMetricItem("ip_traffic_out", serverconfigs.MetricItemCategoryHTTP, "独立IP下行流量", []string{"${remoteAddr}"}, 1, "day", "${countTrafficOut}", []maps.Map{
			{
				"name":     "独立IP排行",
				"type":     "bar",
				"widthDiv": 0,
				"code":     "ip_traffic_out_bar",
			},
		})
		if err != nil {
			return err
		}
	}

	{
		err := createMetricItem("request_path", serverconfigs.MetricItemCategoryHTTP, "请求路径统计", []string{"${requestPath}"}, 1, "day", "${countRequest}", []maps.Map{
			{
				"name":     "请求路径排行",
				"type":     "bar",
				"widthDiv": 0,
				"code":     "request_path_bar",
			},
		})
		if err != nil {
			return err
		}
	}

	{
		err := createMetricItem("request_method", serverconfigs.MetricItemCategoryHTTP, "请求方法统计", []string{"${requestMethod}"}, 1, "day", "${countRequest}", []maps.Map{
			{
				"name":     "请求方法分布",
				"type":     "pie",
				"widthDiv": 2,
				"code":     "request_method_pie",
			},
		})
		if err != nil {
			return err
		}
	}

	{
		err := createMetricItem("status", serverconfigs.MetricItemCategoryHTTP, "状态码统计", []string{"${status}"}, 1, "day", "${countRequest}", []maps.Map{
			{
				"name":     "状态码分布",
				"type":     "pie",
				"widthDiv": 2,
				"code":     "status_pie",
			},
		})
		if err != nil {
			return err
		}
	}

	{
		err := createMetricItem("request_referer_host", serverconfigs.MetricItemCategoryHTTP, "请求来源统计", []string{"${referer.host}"}, 1, "day", "${countRequest}", []maps.Map{
			{
				"name":     "请求来源排行",
				"type":     "bar",
				"widthDiv": 0,
				"code":     "request_referer_host_bar",
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// 更新Agents表
func (this *SQLExecutor) checkClientAgents(db *dbs.DB) error {
	ones, _, err := db.FindOnes("SELECT id FROM edgeClientAgents")
	if err != nil {
		return err
	}

	for _, one := range ones {
		var agentId = one.GetInt64("id")

		countIPs, err := db.FindCol(0, "SELECT COUNT(*) FROM edgeClientAgentIPs WHERE agentId=?", agentId)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE edgeClientAgents SET countIPs=? WHERE id=?", countIPs, agentId)
		if err != nil {
			return err
		}
	}

	return nil
}

// 更新版本号
func (this *SQLExecutor) updateVersion(db *dbs.DB, version string) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeVersions")
	if err != nil {
		return fmt.Errorf("query version failed: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	col, err := stmt.FindCol(0)
	if err != nil {
		return fmt.Errorf("query version failed: %w", err)
	}
	var count = types.Int(col)
	if count > 0 {
		_, err = db.Exec("UPDATE edgeVersions SET version=?", version)
		if err != nil {
			return fmt.Errorf("update version failed: %w", err)
		}
		return nil
	}

	_, err = db.Exec("INSERT edgeVersions (version) VALUES (?)", version)
	if err != nil {
		return fmt.Errorf("create version failed: %w", err)
	}

	return nil
}
