package models

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strconv"
	"sync"
	"time"
)

var accessLogDBMapping = map[int64]*dbs.DB{} // dbNodeId => DB
var accessLogLocker = &sync.RWMutex{}

type httpAccessLogDefinition struct {
	Name          string
	HasRemoteAddr bool
	HasDomain     bool
	Exists        bool
}

// HTTP服务访问
var httpAccessLogDAOMapping = map[int64]*HTTPAccessLogDAOWrapper{} // dbNodeId => DAO

// HTTPAccessLogDAOWrapper HTTP访问日志DAO
type HTTPAccessLogDAOWrapper struct {
	DAO    *HTTPAccessLogDAO
	NodeId int64
}

func init() {
	initializer := NewDBNodeInitializer()
	dbs.OnReadyDone(func() {
		goman.New(func() {
			initializer.Start()
		})
	})
}

func AllAccessLogDBs() []*dbs.DB {
	accessLogLocker.Lock()
	defer accessLogLocker.Unlock()

	var result = []*dbs.DB{}
	for _, db := range accessLogDBMapping {
		result = append(result, db)
	}

	if len(result) == 0 {
		db, _ := dbs.Default()
		if db != nil {
			result = append(result, db)
		}
	}

	return result
}

// 获取获取DAO
func randomHTTPAccessLogDAO() (dao *HTTPAccessLogDAOWrapper) {
	accessLogLocker.RLock()
	defer accessLogLocker.RUnlock()
	if len(httpAccessLogDAOMapping) == 0 {
		dao = nil
		return
	}

	var daoList = []*HTTPAccessLogDAOWrapper{}

	for _, d := range httpAccessLogDAOMapping {
		daoList = append(daoList, d)
	}

	var l = len(daoList)
	if l == 0 {
		return
	}

	if l == 1 {
		return daoList[0]
	}

	return daoList[rands.Int(0, l-1)]
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
		remotelogs.Error("DB_NODE", err.Error())
	}

	// 定时运行
	ticker := time.NewTicker(60 * time.Second)
	for range ticker.C {
		err := this.loop()
		if err != nil {
			remotelogs.Error("DB_NODE", err.Error())
		}
	}
}

// 单次运行
func (this *DBNodeInitializer) loop() error {
	dbNodes, err := SharedDBNodeDAO.FindAllEnabledAndOnDBNodes(nil)
	if err != nil {
		return err
	}

	var nodeIds = []int64{}
	for _, node := range dbNodes {
		nodeIds = append(nodeIds, int64(node.Id))
	}

	// 关掉老的
	accessLogLocker.Lock()
	var closingDbs = []*dbs.DB{}
	for nodeId, db := range accessLogDBMapping {
		if !lists.ContainsInt64(nodeIds, nodeId) {
			closingDbs = append(closingDbs, db)
			delete(accessLogDBMapping, nodeId)
			delete(httpAccessLogDAOMapping, nodeId)
			delete(nsAccessLogDAOMapping, nodeId)
			remotelogs.Error("DB_NODE", "close db node '"+strconv.FormatInt(nodeId, 10)+"'")
		}
	}
	accessLogLocker.Unlock()
	for _, db := range closingDbs {
		_ = db.Close()
	}

	// 启动新的
	for _, node := range dbNodes {
		var nodeId = int64(node.Id)
		accessLogLocker.Lock()
		db, ok := accessLogDBMapping[nodeId]
		accessLogLocker.Unlock()

		var dsn = node.Username + ":" + node.Password + "@tcp(" + node.Host + ":" + fmt.Sprintf("%d", node.Port) + ")/" + node.Database + "?charset=utf8mb4&timeout=10s"

		if ok {
			// 检查配置是否有变化
			oldConfig, err := db.Config()
			if err != nil {
				remotelogs.Error("DB_NODE", "read database old config failed: "+err.Error())
				continue
			}

			// 如果有变化则关闭
			if oldConfig.Dsn != dsn {
				_ = db.Close()
				db = nil
			}
		}

		if db == nil {
			var config = &dbs.DBConfig{
				Driver: "mysql",
				Dsn:    dsn,
				Prefix: "edge",
			}
			db, err := dbs.NewInstanceFromConfig(config)
			if err != nil {
				remotelogs.Error("DB_NODE", "initialize database config failed: "+err.Error())
				continue
			}

			// 检查表是否存在
			// httpAccessLog
			{
				tableDef, err := SharedHTTPAccessLogManager.FindLastTable(db, timeutil.Format("Ymd"), true)
				if err != nil {
					remotelogs.Error("DB_NODE", "create first table in database node failed: "+err.Error())

					// 创建节点日志
					createLogErr := SharedNodeLogDAO.CreateLog(nil, nodeconfigs.NodeRoleDatabase, nodeId, 0, 0, "error", "ACCESS_LOG", "can not create access log table: "+err.Error(), time.Now().Unix(), "", nil)
					if createLogErr != nil {
						remotelogs.Error("NODE_LOG", createLogErr.Error())
					}

					continue
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
					remotelogs.Error("DB_NODE", "initialize dao failed: "+err.Error())
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

			// 扩展
			initAccessLogDAO(db, node)
		}
	}

	return nil
}
