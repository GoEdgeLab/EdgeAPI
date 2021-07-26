package setup

import (
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"io/ioutil"
	"strings"
	"time"
)

var LatestSQLResult = &SQLDumpResult{}

// SQLExecutor 安装或升级SQL执行器
type SQLExecutor struct {
	dbConfig *dbs.DBConfig
}

func NewSQLExecutor(dbConfig *dbs.DBConfig) *SQLExecutor {
	return &SQLExecutor{
		dbConfig: dbConfig,
	}
}

func NewSQLExecutorFromCmd() (*SQLExecutor, error) {
	// 执行SQL
	config := &dbs.Config{}
	configData, err := ioutil.ReadFile(Tea.ConfigFile("db.yaml"))
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, err
	}
	return NewSQLExecutor(config.DBs[Tea.Env]), nil
}

func (this *SQLExecutor) Run() error {
	db, err := dbs.NewInstanceFromConfig(this.dbConfig)
	if err != nil {
		return err
	}

	sqlDump := NewSQLDump()
	_, err = sqlDump.Apply(db, LatestSQLResult)
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
	// 检查管理员
	err := this.checkAdmin(db)
	if err != nil {
		return err
	}

	// 检查管理员平台节点
	err = this.checkAdminNode(db)
	if err != nil {
		return err
	}

	// 检查用户平台节点
	err = this.checkUserNode(db)
	if err != nil {
		return err
	}

	// 检查集群配置
	err = this.checkCluster(db)
	if err != nil {
		return err
	}

	// 检查IP名单
	err = this.checkIPList(db)
	if err != nil {
		return err
	}

	// 检查指标设置
	err = this.checkMetricItems(db)
	if err != nil {
		return err
	}

	// 更新版本号
	err = this.updateVersion(db, teaconst.Version)
	if err != nil {
		return err
	}

	return nil
}

// 检查管理员
func (this *SQLExecutor) checkAdmin(db *dbs.DB) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeAdmins")
	if err != nil {
		return errors.New("check admin failed: " + err.Error())
	}
	defer func() {
		_ = stmt.Close()
	}()
	col, err := stmt.FindCol(0)
	if err != nil {
		return errors.New("check admin failed: " + err.Error())
	}
	count := types.Int(col)
	if count == 0 {
		_, err = db.Exec("INSERT INTO edgeAdmins (username, password, fullname, isSuper, createdAt, state) VALUES (?, ?, ?, ?, ?, ?)", "admin", stringutil.Md5("123456"), "管理员", 1, time.Now().Unix(), 1)
		if err != nil {
			return errors.New("create admin failed: " + err.Error())
		}
	}
	return nil
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
	count := types.Int(col)
	if count > 0 {
		return nil
	}

	nodeId := rands.HexString(32)
	secret := rands.String(32)
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
	count := types.Int(col)
	if count > 0 {
		return nil
	}

	nodeId := rands.HexString(32)
	secret := rands.String(32)
	_, err = db.Exec("INSERT INTO edgeAPITokens (nodeId, secret, role) VALUES (?, ?, ?)", nodeId, secret, "user")
	if err != nil {
		return err
	}

	return nil
}

// 检查集群配置
func (this *SQLExecutor) checkCluster(db *dbs.DB) error {
	/// 检查是否有集群数字
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeNodeClusters")
	if err != nil {
		return errors.New("query clusters failed: " + err.Error())
	}
	defer func() {
		_ = stmt.Close()
	}()

	col, err := stmt.FindCol(0)
	if err != nil {
		return errors.New("query clusters failed: " + err.Error())
	}
	count := types.Int(col)
	if count > 0 {
		return nil
	}

	// 创建默认集群
	_, err = db.Exec("INSERT INTO edgeNodeClusters (name, useAllAPINodes, state, uniqueId, secret) VALUES (?, ?, ?, ?, ?)", "默认集群", 1, 1, rands.HexString(32), rands.String(32))
	if err != nil {
		return err
	}

	return nil
}

// 检查IP名单
func (this *SQLExecutor) checkIPList(db *dbs.DB) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeIPLists")
	if err != nil {
		return errors.New("query ip lists failed: " + err.Error())
	}
	defer func() {
		_ = stmt.Close()
	}()

	col, err := stmt.FindCol(0)
	if err != nil {
		return errors.New("query ip lists failed: " + err.Error())
	}
	count := types.Int(col)
	if count > 0 {
		return nil
	}

	// 创建名单
	_, err = db.Exec("INSERT INTO edgeIPLists(name, type, code, isPublic, createdAt) VALUES (?, ?, ?, ?, ?)", "公共黑名单", "black", "black", 1, time.Now().Unix())
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO edgeIPLists(name, type, code, isPublic, createdAt) VALUES (?, ?, ?, ?, ?)", "公共白名单", "white", "white", 1, time.Now().Unix())
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

		var itemId int64 = 0
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

		itemId = itemMap.GetInt64("id")

		// chart
		for _, chartMap := range chartMaps {
			chartCode := chartMap.GetString("code")
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

// 更新版本号
func (this *SQLExecutor) updateVersion(db *dbs.DB, version string) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM edgeVersions")
	if err != nil {
		return errors.New("query version failed: " + err.Error())
	}
	defer func() {
		_ = stmt.Close()
	}()

	col, err := stmt.FindCol(0)
	if err != nil {
		return errors.New("query version failed: " + err.Error())
	}
	count := types.Int(col)
	if count > 0 {
		_, err = db.Exec("UPDATE edgeVersions SET version=?", version)
		if err != nil {
			return errors.New("update version failed: " + err.Error())
		}
		return nil
	}

	_, err = db.Exec("INSERT edgeVersions (version) VALUES (?)", version)
	if err != nil {
		return errors.New("create version failed: " + err.Error())
	}

	return nil
}

// 判断某个错误是否可以忽略
func (this *SQLExecutor) canIgnoreError(err error) bool {
	if err == nil {
		return true
	}

	// Error 1050: Table 'xxx' already exists
	if strings.Contains(err.Error(), "Error 1050") {
		return true
	}

	return false
}
