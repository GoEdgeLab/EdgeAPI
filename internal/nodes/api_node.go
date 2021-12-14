package nodes

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/accesslogs"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/events"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"github.com/iwind/gosock/pkg/gosock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

var sharedAPIConfig *configs.APIConfig = nil

type APINode struct {
	serviceInstanceMap    map[string]interface{}
	serviceInstanceLocker sync.Mutex

	sock *gosock.Sock

	isStarting bool
}

func NewAPINode() *APINode {
	return &APINode{
		serviceInstanceMap: map[string]interface{}{},
		sock:               gosock.NewTmpSock(teaconst.ProcessName),
	}
}

func (this *APINode) Start() {
	this.isStarting = true

	logs.Println("[API_NODE]start api node, pid: " + strconv.Itoa(os.Getpid()))

	// 检查数据库连接
	err := this.checkDB()
	if err != nil {
		logs.Println("[API_NODE]" + err.Error())
		return
	}

	// 本地Sock
	logs.Println("[API_NODE]listening sock ...")
	err = this.listenSock()
	if err != nil {
		logs.Println("[API_NODE]" + err.Error())
		return
	}

	// 自动升级
	logs.Println("[API_NODE]auto upgrading ...")
	err = this.autoUpgrade()
	if err != nil {
		logs.Println("[API_NODE]auto upgrade failed: " + err.Error())
		return
	}

	// 自动设置数据库
	logs.Println("[API_NODE]setup database ...")
	err = this.setupDB()
	if err != nil {
		logs.Println("[API_NODE]setup database '" + err.Error() + "'")

		// 不阻断执行
	}

	// 数据库通知启动
	logs.Println("[API_NODE]notify ready ...")
	dbs.NotifyReady()

	// 读取配置
	logs.Println("[API_NODE]reading api config ...")
	config, err := configs.SharedAPIConfig()
	if err != nil {
		logs.Println("[API_NODE]start failed: " + err.Error())
		return
	}
	sharedAPIConfig = config

	// 校验
	apiNode, err := models.SharedAPINodeDAO.FindEnabledAPINodeWithUniqueIdAndSecret(nil, config.NodeId, config.Secret)
	if err != nil {
		logs.Println("[API_NODE]start failed: read api node from database failed: " + err.Error())
		return
	}
	if apiNode == nil {
		logs.Println("[API_NODE]can not start node, wrong 'nodeId' or 'secret'")
		return
	}
	config.SetNumberId(int64(apiNode.Id))

	// 设置rlimit
	_ = utils.SetRLimit(1024 * 1024)

	// 状态变更计时器
	goman.New(func() {
		NewNodeStatusExecutor().Listen()
	})

	// 访问日志存储管理器
	goman.New(func() {
		accesslogs.SharedStorageManager.Start()
	})

	// 监听RPC服务
	remotelogs.Println("API_NODE", "starting RPC server ...")

	isListening := this.listenPorts(apiNode)

	if !isListening {
		remotelogs.Error("API_NODE", "the api node require at least one listening address")
		return
	}

	// 结束启动
	this.isStarting = false

	// 保持进程
	select {}
}

// Daemon 实现守护进程
func (this *APINode) Daemon() {
	path := os.TempDir() + "/edge-api.sock"
	isDebug := lists.ContainsString(os.Args, "debug")
	isDebug = true
	for {
		conn, err := net.DialTimeout("unix", path, 1*time.Second)
		if err != nil {
			if isDebug {
				log.Println("[DAEMON]starting ...")
			}

			// 尝试启动
			err = func() error {
				exe, err := os.Executable()
				if err != nil {
					return err
				}
				cmd := exec.Command(exe)
				err = cmd.Start()
				if err != nil {
					return err
				}
				err = cmd.Wait()
				if err != nil {
					return err
				}
				return nil
			}()

			if err != nil {
				if isDebug {
					log.Println("[DAEMON]", err)
				}
				time.Sleep(1 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		} else {
			_ = conn.Close()
			time.Sleep(5 * time.Second)
		}
	}
}

// InstallSystemService 安装系统服务
func (this *APINode) InstallSystemService() error {
	shortName := teaconst.SystemdServiceName

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	manager := utils.NewServiceManager(shortName, teaconst.ProductName)
	err = manager.Install(exe, []string{})
	if err != nil {
		return err
	}
	return nil
}

// 启动RPC监听
func (this *APINode) listenRPC(listener net.Listener, tlsConfig *tls.Config) error {
	var rpcServer *grpc.Server
	if tlsConfig == nil {
		remotelogs.Println("API_NODE", "listening GRPC http://"+listener.Addr().String()+" ...")
		rpcServer = grpc.NewServer(grpc.MaxRecvMsgSize(128 * 1024 * 1024))
	} else {
		logs.Println("[API_NODE]listening GRPC https://" + listener.Addr().String() + " ...")
		rpcServer = grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)), grpc.MaxRecvMsgSize(128*1024*1024))
	}
	this.registerServices(rpcServer)
	err := rpcServer.Serve(listener)
	if err != nil {
		return errors.New("[API_NODE]start rpc failed: " + err.Error())
	}

	return nil
}

// 检查数据库
func (this *APINode) checkDB() error {
	logs.Println("[API_NODE]checking database connection ...")

	db, err := dbs.Default()
	if err != nil {
		return err
	}

	maxTries := 600
	for i := 0; i <= maxTries; i++ {
		_, err := db.Exec("SELECT 1")
		if err != nil {
			if i == maxTries-1 {
				return err
			} else {
				if i%10 == 0 { // 这让提示不会太多
					logs.Println("[API_NODE]reconnecting to database (" + fmt.Sprintf("%.1f", float32(i*100)/float32(maxTries+1)) + "%) ...")
				}
				time.Sleep(1 * time.Second)
			}
		} else {
			logs.Println("[API_NODE]database connected")
			return nil
		}
	}

	return nil
}

// 自动升级
func (this *APINode) autoUpgrade() error {
	if Tea.IsTesting() {
		return nil
	}

	// 执行SQL
	config := &dbs.Config{}
	configData, err := ioutil.ReadFile(Tea.ConfigFile("db.yaml"))
	if err != nil {
		return errors.New("read database config file failed: " + err.Error())
	}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return errors.New("decode database config failed: " + err.Error())
	}
	dbConfig := config.DBs[Tea.Env]
	db, err := dbs.NewInstanceFromConfig(dbConfig)
	if err != nil {
		return errors.New("load database failed: " + err.Error())
	}
	one, err := db.FindOne("SELECT version FROM edgeVersions LIMIT 1")
	if err != nil {
		return errors.New("query version failed: " + err.Error())
	}
	if one != nil {
		// 如果是同样的版本，则直接认为是最新版本
		version := one.GetString("version")
		if stringutil.VersionCompare(version, teaconst.Version) >= 0 {
			return nil
		}
	}

	// 不使用remotelog()，因为此时还没有启动完成
	logs.Println("[API_NODE]upgrade database starting ...")
	err = setup.NewSQLExecutor(dbConfig).Run(false)
	if err != nil {
		return errors.New("execute sql failed: " + err.Error())
	}
	// 不使用remotelog
	logs.Println("[API_NODE]upgrade database done")
	return nil
}

// 自动设置数据库
func (this *APINode) setupDB() error {
	db, err := dbs.Default()
	if err != nil {
		return err
	}

	// 调整预处理语句数量
	{
		result, err := db.FindOne("SHOW VARIABLES WHERE variable_name='max_prepared_stmt_count'")
		if err != nil {
			return err
		}
		value := result.GetString("Value")
		if regexp.MustCompile(`^\d+$`).MatchString(value) {
			valueInt := types.Int(value)
			if valueInt < 65535 {
				_, err := db.Exec("SET GLOBAL max_prepared_stmt_count=65535")
				if err != nil {
					return errors.New("run 'SET GLOBAL max_prepared_stmt_count' on database failed: " + err.Error() + ", \nyou can change the variable in 'my.cnf': \n~~~\n" + `[mysqld]
max_prepared_stmt_count=65535
~~~
then restart mysqld.`)
				}
			}
		}
	}

	return nil
}

// 启动端口
func (this *APINode) listenPorts(apiNode *models.APINode) (isListening bool) {
	// HTTP
	httpConfig, err := apiNode.DecodeHTTP()
	if err != nil {
		remotelogs.Error("API_NODE", "decode http config: "+err.Error())
		return
	}
	isListening = false
	if httpConfig != nil && httpConfig.IsOn && len(httpConfig.Listen) > 0 {
		for _, listen := range httpConfig.Listen {
			for _, addr := range listen.Addresses() {
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					remotelogs.Error("API_NODE", "listening '"+addr+"' failed: "+err.Error()+", we will try to listen port only")

					// 试着只监听端口
					_, port, err := net.SplitHostPort(addr)
					if err != nil {
						continue
					}
					remotelogs.Println("API_NODE", "retry listening port ':"+port+"' only ...")
					listener, err = net.Listen("tcp", ":"+port)
					if err != nil {
						remotelogs.Error("API_NODE", "listening ':"+port+"' failed: "+err.Error())
						continue
					}
					remotelogs.Println("API_NODE", "retry listening port ':"+port+"' only ok")
				}
				goman.New(func() {
					err := this.listenRPC(listener, nil)
					if err != nil {
						remotelogs.Error("API_NODE", "listening '"+addr+"' rpc: "+err.Error())
						return
					}
				})
				isListening = true
			}
		}
	}

	// HTTPS
	httpsConfig, err := apiNode.DecodeHTTPS(nil, nil)
	if err != nil {
		remotelogs.Error("API_NODE", "decode https config: "+err.Error())
		return
	}
	if httpsConfig != nil &&
		httpsConfig.IsOn &&
		len(httpsConfig.Listen) > 0 &&
		httpsConfig.SSLPolicy != nil &&
		httpsConfig.SSLPolicy.IsOn &&
		len(httpsConfig.SSLPolicy.Certs) > 0 {
		certs := []tls.Certificate{}
		for _, cert := range httpsConfig.SSLPolicy.Certs {
			certs = append(certs, *cert.CertObject())
		}

		for _, listen := range httpsConfig.Listen {
			for _, addr := range listen.Addresses() {
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					remotelogs.Error("API_NODE", "listening '"+addr+"' failed: "+err.Error()+", we will try to listen port only")
					// 试着只监听端口
					_, port, err := net.SplitHostPort(addr)
					if err != nil {
						continue
					}
					remotelogs.Println("API_NODE", "retry listening port ':"+port+"' only ...")
					listener, err = net.Listen("tcp", ":"+port)
					if err != nil {
						remotelogs.Error("API_NODE", "listening ':"+port+"' failed: "+err.Error())
						continue
					}
					remotelogs.Println("API_NODE", "retry listening port ':"+port+"' only ok")
				}
				goman.New(func() {
					err := this.listenRPC(listener, &tls.Config{
						Certificates: certs,
					})
					if err != nil {
						remotelogs.Error("API_NODE", "listening '"+addr+"' rpc: "+err.Error())
						return
					}
				})
				isListening = true
			}
		}
	}

	// Rest HTTP
	restHTTPConfig, err := apiNode.DecodeRestHTTP()
	if err != nil {
		remotelogs.Error("API_NODE", "decode REST http config: "+err.Error())
		return
	}
	if restHTTPConfig != nil && restHTTPConfig.IsOn && len(restHTTPConfig.Listen) > 0 {
		for _, listen := range restHTTPConfig.Listen {
			for _, addr := range listen.Addresses() {
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					remotelogs.Error("API_NODE", "listening REST 'http://"+addr+"' failed: "+err.Error())
					continue
				}
				goman.New(func() {
					remotelogs.Println("API_NODE", "listening REST http://"+addr+" ...")
					server := &RestServer{}
					err := server.Listen(listener)
					if err != nil {
						remotelogs.Error("API_NODE", "listening REST 'http://"+addr+"' failed: "+err.Error())
						return
					}
				})
				isListening = true
			}
		}
	}

	// Rest HTTPS
	restHTTPSConfig, err := apiNode.DecodeRestHTTPS(nil, nil)
	if err != nil {
		remotelogs.Error("API_NODE", "decode REST https config: "+err.Error())
		return
	}
	if restHTTPSConfig != nil &&
		restHTTPSConfig.IsOn &&
		len(restHTTPSConfig.Listen) > 0 &&
		restHTTPSConfig.SSLPolicy != nil &&
		restHTTPSConfig.SSLPolicy.IsOn &&
		len(restHTTPSConfig.SSLPolicy.Certs) > 0 {
		for _, listen := range restHTTPSConfig.Listen {
			for _, addr := range listen.Addresses() {
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					remotelogs.Error("API_NODE", "listening REST 'https://"+addr+"' failed: "+err.Error())
					continue
				}
				goman.New(func() {
					remotelogs.Println("API_NODE", "listening REST https://"+addr+" ...")
					server := &RestServer{}

					certs := []tls.Certificate{}
					for _, cert := range httpsConfig.SSLPolicy.Certs {
						certs = append(certs, *cert.CertObject())
					}

					err := server.ListenHTTPS(listener, &tls.Config{
						Certificates: certs,
					})
					if err != nil {
						remotelogs.Error("API_NODE", "listening REST 'https://"+addr+"' failed: "+err.Error())
						return
					}
				})
				isListening = true
			}
		}
	}

	return
}

// 监听本地sock
func (this *APINode) listenSock() error {
	// 检查是否在运行
	if this.sock.IsListening() {
		reply, err := this.sock.Send(&gosock.Command{Code: "pid"})
		if err == nil {
			return errors.New("error: the process is already running, pid: " + maps.NewMap(reply.Params).GetString("pid"))
		} else {
			return errors.New("error: the process is already running")
		}
	}

	// 启动监听
	goman.New(func() {
		this.sock.OnCommand(func(cmd *gosock.Command) {
			switch cmd.Code {
			case "pid":
				_ = cmd.Reply(&gosock.Command{
					Code: "pid",
					Params: map[string]interface{}{
						"pid": os.Getpid(),
					},
				})
			case "info":
				exePath, _ := os.Executable()
				_ = cmd.Reply(&gosock.Command{
					Code: "info",
					Params: map[string]interface{}{
						"pid":     os.Getpid(),
						"version": teaconst.Version,
						"path":    exePath,
					},
				})
			case "stop":
				_ = cmd.ReplyOk()

				// 退出主进程
				events.Notify(events.EventQuit)
				os.Exit(0)
			case "starting": // 是否正在启动
				_ = cmd.Reply(&gosock.Command{
					Code: "starting",
					Params: map[string]interface{}{
						"isStarting": this.isStarting,
					},
				})
			case "goman":
				var posMap = map[string]maps.Map{} // file#line => Map
				for _, instance := range goman.List() {
					var pos = instance.File + "#" + types.String(instance.Line)
					m, ok := posMap[pos]
					if ok {
						m["count"] = m["count"].(int) + 1
					} else {
						m = maps.Map{
							"pos":   pos,
							"count": 1,
						}
						posMap[pos] = m
					}
				}

				var result = []maps.Map{}
				for _, m := range posMap {
					result = append(result, m)
				}

				sort.Slice(result, func(i, j int) bool {
					return result[i]["count"].(int) > result[j]["count"].(int)
				})

				_ = cmd.Reply(&gosock.Command{
					Params: map[string]interface{}{
						"total":  runtime.NumGoroutine(),
						"result": result,
					},
				})
			}
		})

		err := this.sock.Listen()
		if err != nil {
			logs.Println("API_NODE", err.Error())
		}
	})

	events.On(events.EventQuit, func() {
		logs.Println("API_NODE", "quit unix sock")
		_ = this.sock.Close()
	})

	return nil
}
