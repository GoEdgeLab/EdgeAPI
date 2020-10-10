package models

import (
	"fmt"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"hash/crc32"
	"strconv"
	"strings"
	"sync"
	"time"
)

var accessLogDBMapping = map[int64]*dbs.DB{}            // dbNodeId => DB
var accessLogDAOMapping = map[int64]*HTTPAccessLogDAO{} // dbNodeId => DAO
var accessLogLocker = &sync.RWMutex{}
var accessLogTableMapping = map[string]bool{} // tableName_crc(dsn) => true

func init() {
	initializer := NewDBNodeInitializer()
	go initializer.Start()
}

// 获取获取DAO
func randomAccessLogDAO() (dao *HTTPAccessLogDAO) {
	accessLogLocker.RLock()
	if len(accessLogDAOMapping) == 0 {
		dao = nil
	} else {
		for _, d := range accessLogDAOMapping {
			dao = d
			break
		}
	}
	accessLogLocker.RUnlock()
	return
}

// 根据日期获取表名
func findAccessLogTable(db *dbs.DB, day string, force bool) (string, error) {
	config, err := db.Config()
	if err != nil {
		return "", err
	}

	tableName := "edgeHTTPAccessLogs_" + day
	cacheKey := tableName + "_" + fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(config.Dsn)))

	if !force {
		accessLogLocker.RLock()
		_, ok := accessLogTableMapping[cacheKey]
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
		accessLogTableMapping[cacheKey] = true
		accessLogLocker.Unlock()
		return tableName, nil
	}

	// 创建表格
	_, err = db.Exec("CREATE TABLE `" + tableName + "` (\n  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n  `serverId` int(11) unsigned DEFAULT '0' COMMENT '服务ID',\n  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n  `status` int(3) unsigned DEFAULT '0' COMMENT '状态码',\n  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n  `content` json DEFAULT NULL COMMENT '日志内容',\n  `day` varchar(8) DEFAULT NULL COMMENT '日期Ymd',\n  PRIMARY KEY (`id`),\n  KEY `serverId` (`serverId`),\n  KEY `nodeId` (`nodeId`),\n  KEY `serverId_status` (`serverId`,`status`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")
	if err != nil {
		return tableName, err
	}

	accessLogLocker.Lock()
	accessLogTableMapping[cacheKey] = true
	accessLogLocker.Unlock()

	return tableName, nil
}

// 初始化数据库连接
type DBNodeInitializer struct {
}

func NewDBNodeInitializer() *DBNodeInitializer {
	return &DBNodeInitializer{}
}

// 启动
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
	dbNodes, err := SharedDBNodeDAO.FindAllEnabledAndOnDBNodes()
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
		if !this.containsInt64(nodeIds, nodeId) {
			closingDbs = append(closingDbs, db)
			delete(accessLogDBMapping, nodeId)
			delete(accessLogDAOMapping, nodeId)
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
			tableName, err := findAccessLogTable(db, timeutil.Format("Ymd"), false)
			if err != nil {
				if !strings.Contains(err.Error(), "1050") { // 非表格已存在错误
					logs.Println("[DB_NODE]create first table in database node failed: " + err.Error())

					// 创建节点日志
					createLogErr := SharedNodeLogDAO.CreateLog(NodeRoleDatabase, nodeId, "error", "ACCESS_LOG", "can not create access log table: "+err.Error(), time.Now().Unix())
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
			accessLogDAOMapping[nodeId] = dao
			accessLogLocker.Unlock()
		}
	}

	return nil
}

// 判断是否包含某数字
func (this *DBNodeInitializer) containsInt64(values []int64, value int64) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
