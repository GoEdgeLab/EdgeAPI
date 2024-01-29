// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package instances

import (
	"encoding/json"
	"errors"
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/installers/helpers"
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	executils "github.com/TeaOSLab/EdgeAPI/internal/utils/exec"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Instance struct {
	options Options
}

func NewInstance(options Options) *Instance {
	return &Instance{
		options: options,
	}
}

func (this *Instance) SetupAll() error {
	type TaskFn func() error
	for _, taskFn := range []TaskFn{
		this.SetupDB,
		this.SetupAdminNode,
		this.SetupAPINode,
		this.SetupNode,
		this.SetupUserNode,
		this.Clean,
		this.Startup,
	} {
		err := taskFn()
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Instance) SetupDB() error {
	if this.options.Verbose {
		log.Println("setup db ...")
	}

	// 检查数据库名
	{
		db, dbErr := this.dbInstanceWithoutDBName()
		if dbErr != nil {
			return dbErr
		}

		defer func() {
			_ = db.Close()
		}()

		_, err := db.Exec("USE `" + this.options.DB.Name + "`")
		if err != nil {
			if models.CheckSQLErrCode(err, 1049) {
				_, err = db.Exec("CREATE DATABASE `" + this.options.DB.Name + "` DEFAULT CHARSET utf8mb4")
				if err != nil {
					return fmt.Errorf("create new database failed: %w", err)
				}
			} else {
				return fmt.Errorf("check database failed: %w", err)
			}
		}
	}

	// 检查版本
	{
		db, instanceErr := this.dbInstance(false)
		if instanceErr != nil {
			return instanceErr
		}

		defer func() {
			_ = db.Close()
		}()

		dbConfig, configErr := db.Config()
		if configErr != nil {
			return configErr
		}

		var shouldExecute bool
		version, err := db.FindCol(0, "SELECT version FROM edgeVersions")
		if err != nil {
			shouldExecute = true
		} else {
			if !strings.HasPrefix(types.String(version), teaconst.Version+".") {
				shouldExecute = true
			}
		}
		if shouldExecute {
			var executor = setup.NewSQLExecutor(dbConfig)
			err = executor.Run(false)
			if err != nil {
				return fmt.Errorf("execute sql failed: %w", err)
			}

			// wait to commit
			time.Sleep(1 * time.Second)
		}
	}

	// 启用数据库
	db, instanceErr := this.dbInstance(true)
	if instanceErr != nil {
		return instanceErr
	}

	defer func() {
		_ = db.Close()
	}()

	// 创建Admin Token
	var tx *dbs.Tx
	{
		one, err := db.FindOne("SELECT * FROM edgeAPITokens WHERE role='admin'")
		if err != nil {
			return err
		}
		if len(one) == 0 {
			var nodeId = rands.HexString(32)
			var secret = rands.String(32)
			err = models.SharedApiTokenDAO.CreateAPIToken(tx, nodeId, secret, "admin")
			if err != nil {
				return fmt.Errorf("create admin node token failed: %w", err)
			}
		}
	}

	// 创建Admin
	{
		one, err := db.FindOne("SELECT * FROM edgeAdmins")
		if err != nil {
			return err
		}
		if len(one) == 0 {
			var password = rands.String(16)
			// 保存密码到文件
			var dir = this.options.WorkDir + "/usr/local/goedge"
			_, err = os.Stat(dir)
			if err != nil {
				err = os.MkdirAll(dir, 0777)
				if err != nil {
					return fmt.Errorf("create directory '"+dir+"' failed: %w", err)
				}
			}
			err = os.WriteFile(dir+"/password.txt", []byte("Admin Password: "+password+"\n"), 0666)
			if err != nil {
				return fmt.Errorf("write 'password.txt' failed: %w", err)
			}

			_, err = models.SharedAdminDAO.CreateAdmin(tx, "admin", true, password, "Admin", true, []byte("[]"))
			if err != nil {
				return fmt.Errorf("create admin failed: %w", err)
			}
		}
	}

	// 创建API Node
	{
		one, findErr := db.FindOne("SELECT * FROM edgeAPINodes")
		if findErr != nil {
			return findErr
		}
		if len(one) == 0 {
			// http
			var httpConfig = &serverconfigs.HTTPProtocolConfig{}
			httpConfig.IsOn = true
			httpConfig.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  "http",
					PortRange: types.String(this.options.APINode.HTTPPort),
				},
			}
			httpJSON, encodeErr := json.Marshal(httpConfig)
			if encodeErr != nil {
				return fmt.Errorf("encode api http config failed: %w", encodeErr)
			}

			// rest
			var restHTTPConfig = &serverconfigs.HTTPProtocolConfig{}
			restHTTPConfig.IsOn = true
			restHTTPConfig.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  "http",
					PortRange: types.String(this.options.APINode.RestHTTPPort),
				},
			}
			restHTTPJSON, encodeErr := json.Marshal(restHTTPConfig)
			if encodeErr != nil {
				return fmt.Errorf("encode api rest http config failed: %w", encodeErr)
			}

			// access addrs
			var accessAddrs = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  "http",
					Host:      "127.0.0.1",
					PortRange: types.String(this.options.APINode.HTTPPort),
				},
			}
			accessAddrsJSON, encodeErr := json.Marshal(accessAddrs)
			if encodeErr != nil {
				return fmt.Errorf("encode access addresses failed: %w", encodeErr)
			}

			_, err := models.SharedAPINodeDAO.CreateAPINode(tx, "API Node", "Primary API Node", httpJSON, nil, true, restHTTPJSON, nil, accessAddrsJSON, true)
			if err != nil {
				return err
			}
		}
		{
			// check token
			nodes, nodesErr := models.SharedAPINodeDAO.FindAllEnabledAPINodes(tx)
			if nodesErr != nil {
				return nodesErr
			}
			for _, node := range nodes {
				token, err := models.SharedApiTokenDAO.FindEnabledTokenWithNode(tx, node.UniqueId)
				if err != nil {
					return err
				}
				if token == nil {
					err = models.SharedApiTokenDAO.CreateAPIToken(tx, node.UniqueId, node.Secret, nodeconfigs.NodeRoleAPI)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// 创建Node
	var clusterId int64
	{
		{
			// check cluster
			clusterIdCol, err := db.FindCol(0, "SELECT id FROM edgeNodeClusters WHERE state=1")
			if err != nil {
				return err
			}
			clusterId = types.Int64(clusterIdCol)
			if clusterId == 0 {
				return errors.New("invalid cluster id '" + types.String(clusterId) + "'")
			}
		}

		one, findErr := db.FindOne("SELECT * FROM edgeNodes")
		if findErr != nil {
			return findErr
		}
		if len(one) == 0 {
			_, err := models.SharedNodeDAO.CreateNode(tx, 0, "Local Node", clusterId, 0, 0)
			if err != nil {
				return fmt.Errorf("create node failed: %w", err)
			}
		}
	}

	// 检查User
	var userId int64
	{
		one, findErr := db.FindOne("SELECT id FROM edgeUsers WHERE state=1")
		if findErr != nil {
			return findErr
		}
		if len(one) == 0 {
			return errors.New("user must not be created yet")
		}
		userId = one.GetInt64("id")
	}

	// 创建用户节点
	{
		one, findErr := db.FindOne("SELECT * FROM edgeUserNodes")
		if findErr != nil {
			return findErr
		}
		if len(one) == 0 {
			// http
			var httpConfig = &serverconfigs.HTTPProtocolConfig{}
			httpConfig.IsOn = true
			httpConfig.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  "http",
					PortRange: types.String(this.options.UserNode.HTTPPort),
				},
			}
			httpJSON, encodeErr := json.Marshal(httpConfig)
			if encodeErr != nil {
				return fmt.Errorf("encode api http config failed: %w", encodeErr)
			}

			// access addrs
			var accessAddrs = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  "http",
					Host:      "127.0.0.1",
					PortRange: types.String(this.options.UserNode.HTTPPort),
				},
			}
			accessAddrsJSON, encodeErr := json.Marshal(accessAddrs)
			if encodeErr != nil {
				return fmt.Errorf("encode access addresses failed: %w", encodeErr)
			}

			_, err := models.SharedUserNodeDAO.CreateUserNode(tx, "User Platform", "Created by system", httpJSON, nil, accessAddrsJSON, true)
			if err != nil {
				return err
			}
		}
	}

	// 创建网站
	{
		one, findErr := db.FindOne("SELECT id FROM edgeServers WHERE state=1")
		if findErr != nil {
			return findErr
		}
		if len(one) == 0 {

			var webId int64

			{
				{
					var err error
					webId, err = models.SharedHTTPWebDAO.CreateWeb(tx, 0, userId, nil)
					if err != nil {
						return fmt.Errorf("create web failed: %w", err)
					}
				}

				// 访问日志
				{
					err := models.SharedHTTPWebDAO.UpdateWebAccessLogConfig(tx, webId, []byte(`{
			"isPrior": false,
			"isOn": true,
			"fields": [1, 2, 6, 7],
			"status1": true,
			"status2": true,
			"status3": true,
			"status4": true,
			"status5": true,

			"storageOnly": false,
			"storagePolicies": [],

            "firewallOnly": false
		}`))
					if err != nil {
						return err
					}
				}

				// websocket
				{
					websocketId, err := models.SharedHTTPWebsocketDAO.CreateWebsocket(tx, []byte(`{
					"count": 30,
					"unit": "second"
				}`), true, nil, true, "")
					if err != nil {
						return err
					}

					err = models.SharedHTTPWebDAO.UpdateWebsocket(tx, webId, []byte(`{
				"isPrior": false,
				"isOn": true,
				"websocketId": `+types.String(websocketId)+`
			}`))
					if err != nil {
						return err
					}
				}

				// cache
				{
					var cacheConfig = &serverconfigs.HTTPCacheConfig{
						IsPrior:         false,
						IsOn:            true,
						AddStatusHeader: true,
						PurgeIsOn:       false,
						PurgeKey:        "",
						CacheRefs:       []*serverconfigs.HTTPCacheRef{},
					}
					cacheConfigJSON, err := json.Marshal(cacheConfig)
					if err != nil {
						return err
					}
					err = models.SharedHTTPWebDAO.UpdateWebCache(tx, webId, cacheConfigJSON)
					if err != nil {
						return err
					}
				}

				// waf
				{
					var firewallRef = &firewallconfigs.HTTPFirewallRef{
						IsPrior:          false,
						IsOn:             true,
						FirewallPolicyId: 0,
					}
					firewallRefJSON, err := json.Marshal(firewallRef)
					if err != nil {
						return err
					}
					err = models.SharedHTTPWebDAO.UpdateWebFirewall(tx, webId, firewallRefJSON)
					if err != nil {
						return err
					}
				}

				// stat
				{
					var statConfig = &serverconfigs.HTTPStatRef{
						IsPrior: false,
						IsOn:    true,
					}
					statJSON, err := json.Marshal(statConfig)
					if err != nil {
						return err
					}
					err = models.SharedHTTPWebDAO.UpdateWebStat(tx, webId, statJSON)
					if err != nil {
						return err
					}
				}
			}

			var httpConfig = &serverconfigs.HTTPProtocolConfig{}
			httpConfig.IsOn = true
			httpConfig.Listen = []*serverconfigs.NetworkAddressConfig{
				{
					Protocol:  "http",
					PortRange: types.String(this.options.Node.HTTPPort),
				},
			}
			httpConfigJSON, err := json.Marshal(httpConfig)
			if err != nil {
				return err
			}

			// reverse proxy
			reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, 0, userId, nil, nil, nil)
			if err != nil {
				return err
			}
			var reverseProxyRef = &serverconfigs.ReverseProxyRef{
				IsOn:           true,
				ReverseProxyId: reverseProxyId,
			}
			reverseProxyRefJSON, err := json.Marshal(reverseProxyRef)
			if err != nil {
				return err
			}

			_, err = models.SharedServerDAO.CreateServer(tx, 0, userId, serverconfigs.ServerTypeHTTPProxy, "First Site", "Created by system", []byte(`[{"name": "example.org", "type":"full"}]`), false, nil, httpConfigJSON, nil, nil, nil, nil, nil, webId, reverseProxyRefJSON, clusterId, nil, nil, nil, 0)
			if err != nil {
				return fmt.Errorf("create server failed: %w", err)
			}
		}
	}

	// 删除任务
	{
		err := models.SharedNodeTaskDAO.DeleteAllNodeTasks(tx)
		if err != nil {
			return err
		}
	}

	// 设置未初始化
	{
		err := models.SharedSysSettingDAO.UpdateSetting(tx, systemconfigs.SettingCodeStandaloneInstanceInitialized, []byte("0"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Instance) SetupAdminNode() error {
	if this.options.Verbose {
		log.Println("setup admin node ...")
	}

	{
		err := this.unzip("admin", teaconst.Version)
		if err != nil {
			return err
		}
	}

	// create api_node.yaml
	db, findErr := this.dbInstance(true)
	if findErr != nil {
		return findErr
	}

	var apiYAMLData []byte
	{
		node, err := db.FindOne("SELECT nodeId,secret FROM edgeAPITokens WHERE state=1 AND role='admin'")
		if err != nil {
			return err
		}
		if node == nil {
			return errors.New("can not find admin api token in database")
		}

		var apiConfig = &APIConfig{
			RPCEndpoints:     []string{"http://127.0.0.1:" + types.String(this.options.APINode.HTTPPort)},
			RPCDisableUpdate: true,
			NodeId:           node.GetString("nodeId"),
			Secret:           node.GetString("secret"),
		}
		data, err := apiConfig.AsYAML()
		if err != nil {
			return err
		}
		apiYAMLData = data
		err = os.WriteFile(this.targetPrefix()+"/edge-admin/configs/api_admin.yaml", data, 0666)
		if err != nil {
			return err
		}
	}

	// backup
	// ignore errors
	{
		homeDir, err := os.UserHomeDir()
		if err == nil {
			var dir = homeDir + "/.edge-admin/"
			_, err = os.Stat(dir)
			if err != nil {
				_ = os.MkdirAll(dir, 0777)
			}
			_ = os.WriteFile(dir+"/api_admin.yaml", apiYAMLData, 0666)
		}
	}

	// create database config
	{
		dbConfig, err := db.Config()
		if err != nil {
			return fmt.Errorf("retrieve database config failed: %w", err)
		}
		var fullConfig = dbs.Config{}
		fullConfig.Default.DB = "prod"
		fullConfig.DBs = map[string]*dbs.DBConfig{
			"prod": dbConfig,
		}

		data, err := yaml.Marshal(fullConfig)
		if err != nil {
			return err
		}
		err = os.WriteFile(this.targetPrefix()+"/edge-admin/configs/api_db.yaml", data, 0666)
		if err != nil {
			return err
		}
	}

	// server.yaml
	{
		var serverConfig = &TeaGo.ServerConfig{}
		serverConfig.Env = "prod"
		serverConfig.Http.On = true
		serverConfig.Http.Listen = []string{":" + types.String(this.options.AdminNode.Port)}
		data, err := yaml.Marshal(serverConfig)
		if err != nil {
			return err
		}
		err = os.WriteFile(this.targetPrefix()+"/edge-admin/configs/server.yaml", data, 0666)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Instance) SetupAPINode() error {
	if this.options.Verbose {
		log.Println("setup api node ...")
	}

	{
		err := this.unzip("api", teaconst.NodeVersion)
		if err != nil {
			return err
		}
	}

	// create api_node.yaml
	db, findErr := this.dbInstance(true)
	if findErr != nil {
		return findErr
	}

	{
		node, err := db.FindOne("SELECT uniqueId,secret FROM edgeAPINodes WHERE state=1")
		if err != nil {
			return err
		}
		if node == nil {
			return errors.New("can not find node in database")
		}

		var apiConfig = &APIConfig{
			NodeId: node.GetString("uniqueId"),
			Secret: node.GetString("secret"),
		}
		data, err := apiConfig.AsYAML()
		if err != nil {
			return err
		}
		err = os.WriteFile(this.targetPrefix()+"/edge-api/configs/api.yaml", data, 0666)
		if err != nil {
			return err
		}
	}

	// create database config
	{
		dbConfig, err := db.Config()
		if err != nil {
			return fmt.Errorf("retrieve database config failed: %w", err)
		}
		var fullConfig = dbs.Config{}
		fullConfig.Default.DB = "prod"
		fullConfig.DBs = map[string]*dbs.DBConfig{
			"prod": dbConfig,
		}

		data, err := yaml.Marshal(fullConfig)
		if err != nil {
			return err
		}
		err = os.WriteFile(this.targetPrefix()+"/edge-api/configs/db.yaml", data, 0666)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Instance) SetupNode() error {
	if this.options.Verbose {
		log.Println("setup node ...")
	}

	{
		err := this.unzip("node", teaconst.NodeVersion)
		if err != nil {
			return err
		}
	}

	// create api_node.yaml
	{
		db, err := this.dbInstance(true)
		if err != nil {
			return err
		}
		node, err := db.FindOne("SELECT uniqueId,secret FROM edgeNodes WHERE state=1")
		if err != nil {
			return err
		}
		if node == nil {
			return errors.New("can not find node in database")
		}

		var apiConfig = &APIConfig{
			RPCEndpoints:     []string{"http://127.0.0.1:" + types.String(this.options.APINode.HTTPPort)},
			RPCDisableUpdate: true,
			NodeId:           node.GetString("uniqueId"),
			Secret:           node.GetString("secret"),
		}
		data, err := apiConfig.AsYAML()
		if err != nil {
			return err
		}
		err = os.WriteFile(this.targetPrefix()+"/edge-node/configs/api_node.yaml", data, 0666)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Instance) SetupUserNode() error {
	if !this.isPlus() {
		return nil
	}

	if this.options.Verbose {
		log.Println("setup user node ...")
	}

	{
		err := this.unzip("user", teaconst.NodeVersion)
		if err != nil {
			return err
		}
	}

	// create api_node.yaml
	{
		db, err := this.dbInstance(true)
		if err != nil {
			return err
		}
		node, err := db.FindOne("SELECT uniqueId,secret FROM edgeUserNodes WHERE state=1")
		if err != nil {
			return err
		}
		if node == nil {
			return errors.New("can not find node in database")
		}

		var apiConfig = &APIConfig{
			RPCEndpoints:     []string{"http://127.0.0.1:" + types.String(this.options.APINode.HTTPPort)},
			RPCDisableUpdate: true,
			NodeId:           node.GetString("uniqueId"),
			Secret:           node.GetString("secret"),
		}
		data, err := apiConfig.AsYAML()
		if err != nil {
			return err
		}
		err = os.WriteFile(this.targetPrefix()+"/edge-user/configs/api_user.yaml", data, 0666)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Instance) Clean() error {
	if this.options.Verbose {
		log.Println("cleaning ...")
	}

	{
		var dir = this.targetPrefix() + "/edge-admin/edge-api"
		_, err := os.Stat(dir)
		if err == nil {
			err = os.RemoveAll(dir)
			if err != nil {
				return fmt.Errorf("remove %s failed: %w", dir, err)
			}
		}
	}

	if !this.options.IsTesting {
		matches, _ := filepath.Glob(this.options.SrcDir + "/*.zip")
		for _, filePath := range matches {
			err := os.Remove(filePath)
			if err != nil {
				return fmt.Errorf("remove %s failed: %w", filePath, err)
			}
		}
	}

	return nil
}

func (this *Instance) Startup() error {
	if this.options.Verbose {
		log.Println("startup ...")
	}

	type NodeInfo struct {
		Role  string
		Ports []int
	}

	for _, node := range []NodeInfo{
		{"api", []int{this.options.APINode.HTTPPort, this.options.APINode.RestHTTPPort}},
		{"admin", []int{this.options.AdminNode.Port}},
		{"node", []int{this.options.Node.HTTPPort}},
		{"user", []int{this.options.UserNode.HTTPPort}},
	} {
		var exe = this.targetPrefix() + "/edge-" + node.Role + "/bin/edge-" + node.Role
		_, err := os.Stat(exe)
		if err != nil {
			continue
		}

		var cmd = executils.NewTimeoutCmd(30*time.Second, exe, "start")
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("start '"+node.Role+"' failed: %w", err)
		}

		for _, port := range node.Ports {
			var isOk bool
			for i := 0; i < 30; /** seconds **/ i++ {
				conn, connErr := net.DialTimeout("tcp", ":"+types.String(port), 5*time.Second)
				if connErr != nil {
					time.Sleep(1 * time.Second)
					continue
				}
				_ = conn.Close()
				isOk = true
			}
			if !isOk {
				return fmt.Errorf("waiting '%s' port '%d' timeout", node.Role, port)
			}
		}
	}

	return nil
}

func (this *Instance) dbInstance(enableDAO bool) (*dbs.DB, error) {
	Tea.Env = "prod"
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    this.options.DB.Username + ":" + this.options.DB.Password + "@tcp(" + this.options.DB.Host + ":" + types.String(this.options.DB.Port) + ")/" + this.options.DB.Name + "?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		return nil, fmt.Errorf("create database instance failed: %w", err)
	}

	if enableDAO {
		this.setupDAO(db)
	}

	return db, nil
}

func (this *Instance) dbInstanceWithoutDBName() (*dbs.DB, error) {
	Tea.Env = "prod"
	db, err := dbs.NewInstanceFromConfig(&dbs.DBConfig{
		Driver: "mysql",
		Dsn:    this.options.DB.Username + ":" + this.options.DB.Password + "@tcp(" + this.options.DB.Host + ":" + types.String(this.options.DB.Port) + ")/?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		return nil, fmt.Errorf("create database instance failed: %w", err)
	}

	return db, nil
}

func (this *Instance) setupDAO(db *dbs.DB) {
	dbConfig, err := db.Config()
	if err == nil {
		dbs.GlobalConfig().DBs = map[string]*dbs.DBConfig{
			"dev":  dbConfig,
			"prod": dbConfig,
		}
	}

	{
		models.SharedNodeClusterDAO = models.NewNodeClusterDAO()
		models.SharedNodeClusterDAO.Instance = db
	}

	{
		models.SharedNodeDAO = models.NewNodeDAO()
		models.SharedNodeDAO.Instance = db
	}

	{
		models.SharedNodeTaskDAO = models.NewNodeTaskDAO()
		models.SharedNodeTaskDAO.Instance = db
	}

	{
		models.SharedSysLockerDAO = models.NewSysLockerDAO()
		models.SharedSysLockerDAO.Instance = db
	}

	{
		models.SharedAdminDAO = models.NewAdminDAO()
		models.SharedAdminDAO.Instance = db
	}

	{
		models.SharedAPINodeDAO = models.NewAPINodeDAO()
		models.SharedAPINodeDAO.Instance = db
	}

	{
		models.SharedApiTokenDAO = models.NewApiTokenDAO()
		models.SharedApiTokenDAO.Instance = db
	}

	{
		models.SharedUserDAO = models.NewUserDAO()
		models.SharedUserDAO.Instance = db
	}

	{
		models.SharedUserNodeDAO = models.NewUserNodeDAO()
		models.SharedUserNodeDAO.Instance = db
	}

	{
		models.SharedServerDAO = models.NewServerDAO()
		models.SharedServerDAO.Instance = db
	}

	{
		models.SharedHTTPWebDAO = models.NewHTTPWebDAO()
		models.SharedHTTPWebDAO.Instance = db
	}

	{
		models.SharedHTTPWebsocketDAO = models.NewHTTPWebsocketDAO()
		models.SharedHTTPWebsocketDAO.Instance = db
	}

	{
		models.SharedHTTPLocationDAO = models.NewHTTPLocationDAO()
		models.SharedHTTPLocationDAO.Instance = db
	}

	{
		models.SharedServerGroupDAO = models.NewServerGroupDAO()
		models.SharedServerGroupDAO.Instance = db
	}

	{
		models.SharedReverseProxyDAO = models.NewReverseProxyDAO()
		models.SharedReverseProxyDAO.Instance = db
	}

	{
		models.SharedSysSettingDAO = models.NewSysSettingDAO()
		models.SharedSysSettingDAO.Instance = db
	}
}

func (this *Instance) unzip(role string, version string) error {
	if !regexp.MustCompile(`^\w+$`).MatchString(role) {
		return errors.New("invalid role '" + role + "'")
	}

	var arch = runtime.GOARCH
	if runtime.GOOS != "linux" {
		arch = "amd64"
	}
	var plusTag string
	if this.isPlus() && lists.ContainsString([]string{nodeconfigs.NodeRoleAdmin, nodeconfigs.NodeRoleAPI, nodeconfigs.NodeRoleNode}, role) {
		plusTag = "-plus"
	}

	var zipFile = this.options.SrcDir + "/edge-" + role + "-linux-" + arch + plusTag + "-v" + version + ".zip"
	{
		stat, err := os.Stat(zipFile)
		if err != nil {
			return fmt.Errorf("stat '"+zipFile+"' failed: %w", err)
		}
		if stat.IsDir() {
			return errors.New("'" + zipFile + "' should be a file instead of directory")
		}
	}

	var targetPrefix = this.targetPrefix()

	{
		stat, err := os.Stat(targetPrefix)
		if err != nil || !stat.IsDir() {
			err = os.MkdirAll(targetPrefix, 0777)
			if err != nil {
				return fmt.Errorf("create directory '"+targetPrefix+"' failed: %w", err)
			}
		}
	}

	if this.options.Cacheable {
		var targetDir = targetPrefix + "/edge-" + role
		_, err := os.Stat(targetDir)
		if err == nil {
			return nil
		}
	}

	var unzip = helpers.NewUnzip(zipFile, targetPrefix)
	return unzip.Run()
}

func (this *Instance) targetPrefix() string {
	return this.options.WorkDir + "/usr/local/goedge"
}

func (this *Instance) isPlus() bool {
	return teaconst.Tag == "plus"
}
