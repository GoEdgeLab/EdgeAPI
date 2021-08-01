package models

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"hash/crc32"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var accessLogDBMapping = map[int64]*dbs.DB{} // dbNodeId => DB
var accessLogLocker = &sync.RWMutex{}

type httpAccessLogDefinition struct {
	Name          string
	HasRemoteAddr bool
	Exists        bool
}

// HTTP服务访问
var httpAccessLogDAOMapping = map[int64]*HTTPAccessLogDAOWrapper{}    // dbNodeId => DAO
var httpAccessLogTableMapping = map[string]*httpAccessLogDefinition{} // tableName_crc(dsn) => true

// DNS服务访问
var nsAccessLogDAOMapping = map[int64]*NSAccessLogDAOWrapper{} // dbNodeId => DAO
var nsAccessLogTableMapping = map[string]bool{}                // tableName_crc(dsn) => true

// HTTPAccessLogDAOWrapper HTTP访问日志DAO
type HTTPAccessLogDAOWrapper struct {
	DAO    *HTTPAccessLogDAO
	NodeId int64
}

// NSAccessLogDAOWrapper NS访问日志DAO
type NSAccessLogDAOWrapper struct {
	DAO    *NSAccessLogDAO
	NodeId int64
}

func init() {
	initializer := NewDBNodeInitializer()
	dbs.OnReadyDone(func() {
		go initializer.Start()
	})
}

// 获取获取DAO
func randomHTTPAccessLogDAO() (dao *HTTPAccessLogDAOWrapper) {
	accessLogLocker.RLock()
	if len(httpAccessLogDAOMapping) == 0 {
		dao = nil
	} else {
		for _, d := range httpAccessLogDAOMapping {
			dao = d
			break
		}
	}
	accessLogLocker.RUnlock()
	return
}

func randomNSAccessLogDAO() (dao *NSAccessLogDAOWrapper) {
	accessLogLocker.RLock()
	if len(nsAccessLogDAOMapping) == 0 {
		dao = nil
	} else {
		for _, d := range nsAccessLogDAOMapping {
			dao = d
			break
		}
	}
	accessLogLocker.RUnlock()
	return
}

// 检查表格是否存在
func findHTTPAccessLogTableName(db *dbs.DB, day string) (tableName string, hasRemoteAddr bool, ok bool, err error) {
	if !regexp.MustCompile(`^\d{8}$`).MatchString(day) {
		err = errors.New("invalid day '" + day + "', should be YYYYMMDD")
		return
	}

	config, err := db.Config()
	if err != nil {
		return "", false, false, err
	}

	tableName = "edgeHTTPAccessLogs_" + day
	cacheKey := tableName + "_" + fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(config.Dsn)))

	accessLogLocker.RLock()
	def, ok := httpAccessLogTableMapping[cacheKey]
	accessLogLocker.RUnlock()
	if ok {
		return tableName, def.HasRemoteAddr, true, nil
	}

	def, err = findHTTPAccessLogTable(db, day, false)
	if err != nil {
		return tableName, false, false, err
	}

	return tableName, def.HasRemoteAddr, def.Exists, nil
}

func findNSAccessLogTableName(db *dbs.DB, day string) (tableName string, ok bool, err error) {
	if !regexp.MustCompile(`^\d{8}$`).MatchString(day) {
		err = errors.New("invalid day '" + day + "', should be YYYYMMDD")
		return
	}

	config, err := db.Config()
	if err != nil {
		return "", false, err
	}

	tableName = "edgeNSAccessLogs_" + day
	cacheKey := tableName + "_" + fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(config.Dsn)))

	accessLogLocker.RLock()
	_, ok = nsAccessLogTableMapping[cacheKey]
	accessLogLocker.RUnlock()
	if ok {
		return tableName, true, nil
	}

	tableNames, err := db.TableNames()
	if err != nil {
		return tableName, false, err
	}

	return tableName, lists.ContainsString(tableNames, tableName), nil
}

// 根据日期获取表名
func findHTTPAccessLogTable(db *dbs.DB, day string, force bool) (*httpAccessLogDefinition, error) {
	config, err := db.Config()
	if err != nil {
		return nil, err
	}

	tableName := "edgeHTTPAccessLogs_" + day
	cacheKey := tableName + "_" + fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(config.Dsn)))

	if !force {
		accessLogLocker.RLock()
		definition, ok := httpAccessLogTableMapping[cacheKey]
		accessLogLocker.RUnlock()
		if ok {
			return definition, nil
		}
	}

	tableNames, err := db.TableNames()
	if err != nil {
		return nil, err
	}

	if lists.ContainsString(tableNames, tableName) {
		table, err := db.FindTable(tableName)
		if err != nil {
			return nil, err
		}

		accessLogLocker.Lock()
		var definition = &httpAccessLogDefinition{
			Name:          tableName,
			HasRemoteAddr: table.FindFieldWithName("remoteAddr") != nil,
			Exists:        true,
		}
		httpAccessLogTableMapping[cacheKey] = definition
		accessLogLocker.Unlock()
		return definition, nil
	}

	if !force {
		return &httpAccessLogDefinition{Name: tableName, HasRemoteAddr: true, Exists: false}, nil
	}

	// 创建表格
	_, err = db.Exec("CREATE TABLE `" + tableName + "` (`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',`serverId` int(11) unsigned DEFAULT '0' COMMENT '服务ID',`nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',`status` int(3) unsigned DEFAULT '0' COMMENT '状态码',`createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',`content` json DEFAULT NULL COMMENT '日志内容',`requestId` varchar(128) DEFAULT NULL COMMENT '请求ID',`firewallPolicyId` int(11) unsigned DEFAULT '0' COMMENT 'WAF策略ID',`firewallRuleGroupId` int(11) unsigned DEFAULT '0' COMMENT 'WAF分组ID',`firewallRuleSetId` int(11) unsigned DEFAULT '0' COMMENT 'WAF集ID',`firewallRuleId` int(11) unsigned DEFAULT '0' COMMENT 'WAF规则ID',`remoteAddr` varchar(64) DEFAULT NULL COMMENT 'IP地址',PRIMARY KEY (`id`),KEY `serverId` (`serverId`),KEY `nodeId` (`nodeId`),KEY `serverId_status` (`serverId`,`status`),KEY `requestId` (`requestId`),KEY `firewallPolicyId` (`firewallPolicyId`),KEY `firewallRuleGroupId` (`firewallRuleGroupId`),KEY `firewallRuleSetId` (`firewallRuleSetId`),	KEY `firewallRuleId` (`firewallRuleId`),	KEY `remoteAddr` (`remoteAddr`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='访问日志';")
	if err != nil {
		return nil, err
	}

	accessLogLocker.Lock()
	var definition = &httpAccessLogDefinition{
		Name:          tableName,
		HasRemoteAddr: true,
		Exists:        true,
	}
	httpAccessLogTableMapping[cacheKey] = definition
	accessLogLocker.Unlock()

	return definition, nil
}

func findNSAccessLogTable(db *dbs.DB, day string, force bool) (string, error) {
	config, err := db.Config()
	if err != nil {
		return "", err
	}

	tableName := "edgeNSAccessLogs_" + day
	cacheKey := tableName + "_" + fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(config.Dsn)))

	if !force {
		accessLogLocker.RLock()
		_, ok := nsAccessLogTableMapping[cacheKey]
		accessLogLocker.RUnlock()
		if ok {
			return tableName, nil
		}
	}

	tableNames, err := db.TableNames()
	if err != nil {
		return tableName, err
	}

	if lists.ContainsString(tableNames, tableName) {
		accessLogLocker.Lock()
		nsAccessLogTableMapping[cacheKey] = true
		accessLogLocker.Unlock()
		return tableName, nil
	}

	// 创建表格
	_, err = db.Exec("CREATE TABLE `" + tableName + "` (\n  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n  `domainId` int(11) unsigned DEFAULT '0' COMMENT '域名ID',\n  `recordId` int(11) unsigned DEFAULT '0' COMMENT '记录ID',\n  `content` json DEFAULT NULL COMMENT '访问数据',\n  `requestId` varchar(128) DEFAULT NULL COMMENT '请求ID',\n  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n  PRIMARY KEY (`id`),\n  KEY `nodeId` (`nodeId`),\n  KEY `domainId` (`domainId`),\n  KEY `recordId` (`recordId`),\n  KEY `requestId` (`requestId`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='域名服务访问日志';")
	if err != nil {
		return tableName, err
	}

	accessLogLocker.Lock()
	nsAccessLogTableMapping[cacheKey] = true
	accessLogLocker.Unlock()

	return tableName, nil
}

// DBNodeInitializer 初始化数据库连接
type DBNodeInitializer struct {
}

func NewDBNodeInitializer() *DBNodeInitializer {
	return &DBNodeInitializer{}
}

// Start 启动
func (this *DBNodeInitializer) Start() {
	// 初始运行
	err := this.loop()
	if err != nil {
		logs.Println("[DB_NODE]" + err.Error())
	}

	// 定时运行
	ticker := time.NewTicker(60 * time.Second)
	for range ticker.C {
		err := this.loop()
		if err != nil {
			logs.Println("[DB_NODE]" + err.Error())
		}
	}
}

// 单次运行
func (this *DBNodeInitializer) loop() error {
	dbNodes, err := SharedDBNodeDAO.FindAllEnabledAndOnDBNodes(nil)
	if err != nil {
		return err
	}

	nodeIds := []int64{}
	for _, node := range dbNodes {
		nodeIds = append(nodeIds, int64(node.Id))
	}

	// 关掉老的
	accessLogLocker.Lock()
	closingDbs := []*dbs.DB{}
	for nodeId, db := range accessLogDBMapping {
		if !lists.ContainsInt64(nodeIds, nodeId) {
			closingDbs = append(closingDbs, db)
			delete(accessLogDBMapping, nodeId)
			delete(httpAccessLogDAOMapping, nodeId)
			delete(nsAccessLogDAOMapping, nodeId)
			logs.Println("[DB_NODE]close db node '" + strconv.FormatInt(nodeId, 10) + "'")
		}
	}
	accessLogLocker.Unlock()
	for _, db := range closingDbs {
		_ = db.Close()
	}

	// 启动新的
	for _, node := range dbNodes {
		nodeId := int64(node.Id)
		accessLogLocker.Lock()
		db, ok := accessLogDBMapping[nodeId]
		accessLogLocker.Unlock()

		dsn := node.Username + ":" + node.Password + "@tcp(" + node.Host + ":" + fmt.Sprintf("%d", node.Port) + ")/" + node.Database + "?charset=utf8mb4&timeout=10s"

		if ok {
			// 检查配置是否有变化
			oldConfig, err := db.Config()
			if err != nil {
				logs.Println("[DB_NODE]read database old config failed: " + err.Error())
				continue
			}

			// 如果有变化则关闭
			if oldConfig.Dsn != dsn {
				_ = db.Close()
				db = nil
			}
		}

		if db == nil {
			config := &dbs.DBConfig{
				Driver: "mysql",
				Dsn:    dsn,
				Prefix: "edge",
			}
			db, err := dbs.NewInstanceFromConfig(config)
			if err != nil {
				logs.Println("[DB_NODE]initialize database config failed: " + err.Error())
				continue
			}

			// 检查表是否存在
			// httpAccessLog
			{
				tableDef, err := findHTTPAccessLogTable(db, timeutil.Format("Ymd"), true)
				if err != nil {
					if !strings.Contains(err.Error(), "1050") { // 非表格已存在错误
						logs.Println("[DB_NODE]create first table in database node failed: " + err.Error())

						// 创建节点日志
						createLogErr := SharedNodeLogDAO.CreateLog(nil, nodeconfigs.NodeRoleDatabase, nodeId, 0, 0, "error", "ACCESS_LOG", "can not create access log table: "+err.Error(), time.Now().Unix())
						if createLogErr != nil {
							logs.Println("[NODE_LOG]" + createLogErr.Error())
						}

						continue
					} else {
						err = nil
					}
				}

				daoObject := dbs.DAOObject{
					Instance: db,
					DB:       node.Name + "(id:" + strconv.Itoa(int(node.Id)) + ")",
					Table:    tableDef.Name,
					PkName:   "id",
					Model:    new(HTTPAccessLog),
				}
				err = daoObject.Init()
				if err != nil {
					logs.Println("[DB_NODE]initialize dao failed: " + err.Error())
					continue
				}

				accessLogLocker.Lock()
				accessLogDBMapping[nodeId] = db
				dao := &HTTPAccessLogDAO{
					DAOObject: daoObject,
				}
				httpAccessLogDAOMapping[nodeId] = &HTTPAccessLogDAOWrapper{
					DAO:    dao,
					NodeId: nodeId,
				}
				accessLogLocker.Unlock()
			}

			// nsAccessLog
			{
				tableName, err := findNSAccessLogTable(db, timeutil.Format("Ymd"), false)
				if err != nil {
					if !strings.Contains(err.Error(), "1050") { // 非表格已存在错误
						logs.Println("[DB_NODE]create first table in database node failed: " + err.Error())

						// 创建节点日志
						createLogErr := SharedNodeLogDAO.CreateLog(nil, nodeconfigs.NodeRoleDatabase, nodeId, 0, 0, "error", "ACCESS_LOG", "can not create access log table: "+err.Error(), time.Now().Unix())
						if createLogErr != nil {
							logs.Println("[NODE_LOG]" + createLogErr.Error())
						}

						continue
					} else {
						err = nil
					}
				}

				daoObject := dbs.DAOObject{
					Instance: db,
					DB:       node.Name + "(id:" + strconv.Itoa(int(node.Id)) + ")",
					Table:    tableName,
					PkName:   "id",
					Model:    new(NSAccessLog),
				}
				err = daoObject.Init()
				if err != nil {
					logs.Println("[DB_NODE]initialize dao failed: " + err.Error())
					continue
				}

				accessLogLocker.Lock()
				accessLogDBMapping[nodeId] = db
				dao := &NSAccessLogDAO{
					DAOObject: daoObject,
				}
				nsAccessLogDAOMapping[nodeId] = &NSAccessLogDAOWrapper{
					DAO:    dao,
					NodeId: nodeId,
				}
				accessLogLocker.Unlock()
			}
		}
	}

	return nil
}
