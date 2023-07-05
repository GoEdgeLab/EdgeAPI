package nodes

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/events"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc"
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/iplibrary"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/go-sql-driver/mysql"
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
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	// grpc decompression
	_ "google.golang.org/grpc/encoding/gzip"
)

var sharedAPIConfig *configs.APIConfig = nil

type APINode struct {
	serviceInstanceMap    map[string]any
	serviceInstanceLocker sync.Mutex

	sock *gosock.Sock

	isStarting bool

	issues     []*StartIssue
	issuesFile string

	progress *utils.Progress
}

func NewAPINode() *APINode {
	return &APINode{
		serviceInstanceMap: map[string]any{},
		sock:               gosock.NewTmpSock(teaconst.ProcessName),

		issues:     []*StartIssue{},
		issuesFile: Tea.LogFile("issues.log"),
	}
}

func (this *APINode) Start() {
	this.isStarting = true

	logs.Println("[API_NODE]start api node, pid: " + strconv.Itoa(os.Getpid()))

	// 保存启动过程中的问题，以便于查看
	defer func() {
		this.saveIssues()
	}()

	// 本地Sock
	logs.Println("[API_NODE]listening sock ...")
	err := this.listenSock()
	if err != nil {
		var errString = "start local sock failed: " + err.Error()
		logs.Println("[API_NODE]" + errString)
		this.addStartIssue("sock", errString, "")
		return
	}

	// 启动IP库
	this.setProgress("IP_LIBRARY", "开始初始化IP库")
	remotelogs.Println("API_NODE", "initializing ip library ...")
	err = iplibrary.InitDefault()
	if err != nil {
		remotelogs.Error("API_NODE", "initialize ip library failed: "+err.Error())
	}

	// 检查数据库连接
	this.setProgress("DATABASE", "正在检查数据库连接")
	err = this.checkDB()
	if err != nil {
		var errString = "check database connection failed: " + err.Error()
		logs.Println("[API_NODE]" + errString)
		this.addStartIssue("db", errString, this.dbIssueSuggestion(err.Error()))
		return
	}

	// 自动升级
	logs.Println("[API_NODE]auto upgrading ...")
	this.setProgress("DATABASE", "正在升级数据库")
	err = this.autoUpgrade()
	if err != nil {
		var errString = "auto upgrade failed: " + err.Error()
		logs.Println("[API_NODE]" + errString)
		this.addStartIssue("db", errString, this.dbIssueSuggestion(err.Error()))
		return
	}

	// 自动设置数据库
	this.setProgress("DATABASE", "正在设置数据库")
	logs.Println("[API_NODE]setup database ...")
	err = this.setupDB()
	if err != nil {
		logs.Println("[API_NODE]setup database '" + err.Error() + "'")

		// 不阻断执行
	}

	// 数据库通知启动
	this.setProgress("DATABASE", "正在建立数据库模型")
	logs.Println("[API_NODE]notify ready ...")
	dbs.NotifyReady()

	// 设置时区
	this.setProgress("TIMEZONE", "正在设置时区")
	this.setupTimeZone()

	// 读取配置
	this.setProgress("DATABASE", "正在加载API配置")
	logs.Println("[API_NODE]reading api config ...")
	config, err := configs.SharedAPIConfig()
	if err != nil {
		var errString = "read api config failed: " + err.Error()
		logs.Println("[API_NODE]" + errString)
		this.addStartIssue("config", errString, "")
		return
	}
	sharedAPIConfig = config

	// 校验
	apiNode, err := models.SharedAPINodeDAO.FindEnabledAPINodeWithUniqueIdAndSecret(nil, config.NodeId, config.Secret)
	if err != nil {
		var errString = "start failed: read api node from database failed: " + err.Error()
		logs.Println("[API_NODE]" + errString)
		this.addStartIssue("db", errString, "")
		return
	}
	if apiNode == nil {
		var errString = "can not start node, wrong 'nodeId' or 'secret'"
		logs.Println("[API_NODE]" + errString)
		this.addStartIssue("config", errString, "请在api.yaml配置文件中填写正确的`nodeId`和`secret`，如果数据库或者管理节点或API节点是从别的服务器迁移过来的，请将老的系统配置拷贝到当前节点配置下")
		return
	}
	config.SetNumberId(int64(apiNode.Id))

	// 清除上一次启动错误
	// 这个错误文件可能不存在，不需要处理错误
	_ = os.Remove(this.issuesFile)

	// 设置rlimit
	_ = utils.SetRLimit(1024 * 1024)

	// 状态变更计时器
	goman.New(func() {
		NewNodeStatusExecutor().Listen()
	})

	// 访问日志存储管理器
	this.setProgress("ACCESS_LOG_STORAGES", "正在启动访问日志存储器")
	this.startAccessLogStorages()

	// 监听RPC服务
	this.setProgress("LISTEN_PORT", "正在启动监听端口")
	remotelogs.Println("API_NODE", "starting RPC server ...")

	var isListening = this.listenPorts(apiNode)

	if !isListening {
		var errString = "the api node require at least one listening address"
		remotelogs.Error("API_NODE", errString)
		this.addStartIssue("config", errString, "请给当前API节点设置一个监听端口")
		return
	}

	// 结束启动
	this.isStarting = false
	this.progress = nil

	// 保持进程
	select {}
}

// Daemon 实现守护进程
func (this *APINode) Daemon() {
	var path = os.TempDir() + "/" + teaconst.ProcessName + ".sock"
	var isDebug = lists.ContainsString(os.Args, "debug")
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
	var shortName = teaconst.SystemdServiceName

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	var manager = utils.NewServiceManager(shortName, teaconst.ProductName)
	err = manager.Install(exe, []string{})
	if err != nil {
		return err
	}
	return nil
}

// 启动RPC监听
func (this *APINode) listenRPC(listener net.Listener, tlsConfig *tls.Config) error {
	var rpcServer *grpc.Server
	var options = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(512 << 20),
		grpc.MaxSendMsgSize(512 << 20),
		grpc.UnaryInterceptor(this.unaryInterceptor),
	}

	if tlsConfig == nil {
		remotelogs.Println("API_NODE", "listening GRPC http://"+listener.Addr().String()+" ...")
		rpcServer = grpc.NewServer(options...)
	} else {
		logs.Println("[API_NODE]listening GRPC https://" + listener.Addr().String() + " ...")
		options = append(options, grpc.Creds(credentials.NewTLS(tlsConfig)))
		rpcServer = grpc.NewServer(options...)
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

	// lookup mysqld_safe process
	go dbutils.FindMySQLPathAndRemember()

	db, err := dbs.Default()
	if err != nil {
		return err
	}

	// 第一次测试连接
	_, err = db.Exec("SELECT 1")
	if err != nil {
		var errString = "check database connection failed: " + err.Error()
		logs.Println("[API_NODE]" + errString)
		this.addStartIssue("db", errString, this.dbIssueSuggestion(errString))

		// 决定是否尝试启动本地的MySQL
		if strings.Contains(err.Error(), "connection refused") {
			config, _ := db.Config()
			if config != nil && (strings.Contains(config.Dsn, "tcp(127.0.0.1:") || strings.Contains(config.Dsn, "tcp(localhost:")) && os.Getgid() == 0 /** ROOT 用户 **/ {
				dbutils.StartLocalMySQL()
			}
		}

		// 多次尝试
		var maxTries = 600
		if Tea.IsTesting() {
			maxTries = 600
		}
		for i := 0; i <= maxTries; i++ {
			_, err = db.Exec("SELECT 1")
			if err != nil {
				if i == maxTries-1 {
					return err
				} else {
					if i%10 == 0 { // 这让提示不会太多
						logs.Println("[API_NODE]check database connection failed: " + err.Error() + ", reconnecting to database ...")
					}
					time.Sleep(1 * time.Second)
				}
			} else {
				logs.Println("[API_NODE]database connected")
				return nil
			}
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
	var config = &dbs.Config{}
	configData, err := os.ReadFile(Tea.ConfigFile("db.yaml"))
	if err != nil {
		return errors.New("read database config file failed: " + err.Error())
	}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return errors.New("decode database config failed: " + err.Error())
	}
	var dbConfig = config.DBs[Tea.Env]
	db, err := dbs.NewInstanceFromConfig(dbConfig)
	if err != nil {
		return errors.New("load database failed: " + err.Error())
	}
	defer func() {
		_ = db.Close()
	}()
	one, err := db.FindOne("SELECT version FROM edgeVersions LIMIT 1")
	if err != nil {
		return errors.New("query version failed: " + err.Error())
	}
	if one != nil {
		// 如果是同样的版本，则直接认为是最新版本
		var version = one.GetString("version")
		if stringutil.VersionCompare(version, setup.ComposeSQLVersion()) >= 0 {
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

	// 检查是否为root用户
	config, _ := db.Config()
	if config == nil {
		return nil
	}
	dsnConfig, err := mysql.ParseDSN(config.Dsn)
	if err != nil || dsnConfig == nil {
		return err
	}
	if dsnConfig.User != "root" {
		return nil
	}

	// 设置Innodb事务提交模式
	{
		result, err := db.FindOne("SHOW VARIABLES WHERE variable_name='innodb_flush_log_at_trx_commit'")
		if err == nil && result != nil {
			var oldValue = result.GetInt("Value")
			if oldValue == 1 {
				_, _ = db.Exec("SET GLOBAL innodb_flush_log_at_trx_commit=2")
			}
		}
	}

	// 调整预处理语句数量
	_ = dbutils.SetGlobalVarMin(db, "max_prepared_stmt_count", 65535)

	// 调整binlog过期时间
	{
		const binlogExpireDays = 7

		version, err := db.FindCol(0, "SELECT VERSION()")
		if err == nil {
			var versionString = types.String(version)
			if strings.HasPrefix(versionString, "8.") {
				_ = dbutils.SetGlobalVarMax(db, "binlog_expire_logs_seconds", binlogExpireDays*86400)
			} else if strings.HasPrefix(versionString, "5.") {
				_ = dbutils.SetGlobalVarMax(db, "expire_logs_days", binlogExpireDays)
			}
		}
	}

	// 设置binlog_cache_size
	_ = dbutils.SetGlobalVarMin(db, "binlog_cache_size", 1*1024*1024)

	// 设置binlog_stmt_cache_size
	_ = dbutils.SetGlobalVarMin(db, "binlog_stmt_cache_size", 1*1024*1024)

	// 设置thread_cache_size
	_ = dbutils.SetGlobalVarMin(db, "thread_cache_size", 32)

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
	var ports = []int{}
	isListening = false
	if httpConfig != nil && httpConfig.IsOn && len(httpConfig.Listen) > 0 {
		for _, listen := range httpConfig.Listen {
			for _, addr := range listen.Addresses() {
				// 收集Port
				_, portString, _ := net.SplitHostPort(addr)
				var port = types.Int(portString)
				if port > 0 && !lists.ContainsInt(ports, port) {
					ports = append(ports, port)
				}

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
				// 收集Port
				_, portString, _ := net.SplitHostPort(addr)
				var port = types.Int(portString)
				if port > 0 && !lists.ContainsInt(ports, port) {
					ports = append(ports, port)
				}

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
				// 收集Port
				_, portString, _ := net.SplitHostPort(addr)
				var port = types.Int(portString)
				if port > 0 && !lists.ContainsInt(ports, port) {
					ports = append(ports, port)
				}

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
				// 收集Port
				_, portString, _ := net.SplitHostPort(addr)
				var port = types.Int(portString)
				if port > 0 && !lists.ContainsInt(ports, port) {
					ports = append(ports, port)
				}

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

	// add to local firewall
	if len(ports) > 0 {
		go utils.AddPortsToFirewall(ports)
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
			case "pid": // 查询PID
				_ = cmd.Reply(&gosock.Command{
					Code: "pid",
					Params: map[string]any{
						"pid": os.Getpid(),
					},
				})
			case "info": // 进程相关信息
				exePath, _ := os.Executable()
				_ = cmd.Reply(&gosock.Command{
					Code: "info",
					Params: map[string]any{
						"pid":     os.Getpid(),
						"version": teaconst.Version,
						"path":    exePath,
					},
				})
			case "stop": // 停止
				_ = cmd.ReplyOk()

				// 退出主进程
				events.Notify(events.EventQuit)
				os.Exit(0)
			case "starting": // 是否正在启动
				_ = cmd.Reply(&gosock.Command{
					Code: "starting",
					Params: map[string]any{
						"isStarting": this.isStarting,
						"progress":   this.progress,
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
					Params: map[string]any{
						"total":  runtime.NumGoroutine(),
						"result": result,
					},
				})
			case "debug": // 进入|取消调试模式
				teaconst.Debug = !teaconst.Debug
				_ = cmd.Reply(&gosock.Command{
					Params: map[string]any{"debug": teaconst.Debug},
				})
			case "db.stmt.prepare": // 显示prepared的语句
				dbs.ShowPreparedStatements = !dbs.ShowPreparedStatements
				_ = cmd.Reply(&gosock.Command{
					Params: map[string]any{"isOn": dbs.ShowPreparedStatements},
				})
			case "db.stmt.count": // 查询prepared语句数量
				db, _ := dbs.Default()
				if db != nil {
					_ = cmd.Reply(&gosock.Command{
						Params: map[string]any{"count": db.StmtManager().Len()},
					})
				} else {
					_ = cmd.Reply(&gosock.Command{
						Params: map[string]any{"count": 0},
					})
				}
			case "instance": // 获取实例代号
				_ = cmd.Reply(&gosock.Command{
					Params: map[string]any{
						"code": teaconst.InstanceCode,
					},
				})
			case "lookupToken":
				var role = maps.NewMap(cmd.Params).GetString("role")
				switch role {
				case "admin", "user", "api":
					tokens, err := models.SharedApiTokenDAO.FindAllEnabledAPITokens(nil, role)
					if err != nil {
						_ = cmd.Reply(&gosock.Command{
							Params: map[string]any{
								"isOk": false,
								"err":  err.Error(),
							},
						})
					} else {
						var tokenMaps = []maps.Map{}
						for _, token := range tokens {
							tokenMaps = append(tokenMaps, maps.Map{
								"nodeId": token.NodeId,
								"secret": token.Secret,
							})
						}
						_ = cmd.Reply(&gosock.Command{
							Params: map[string]any{
								"isOk":   true,
								"tokens": tokenMaps,
							},
						})
					}
				default:
					_ = cmd.Reply(&gosock.Command{
						Params: map[string]any{
							"isOk": false,
							"err":  "unsupported role '" + role + "'",
						},
					})
				}
			}
		})

		err := this.sock.Listen()
		if err != nil {
			remotelogs.Println("API_NODE", err.Error())
		}
	})

	events.On(events.EventQuit, func() {
		remotelogs.Println("API_NODE", "quit unix sock")
		_ = this.sock.Close()
	})

	return nil
}

// 服务过滤器
func (this *APINode) unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if teaconst.Debug {
		var before = time.Now()
		var traceCtx = rpc.NewContext(ctx)
		resp, err = handler(traceCtx, req)

		var costMs = time.Since(before).Seconds() * 1000
		statErr := models.SharedAPIMethodStatDAO.CreateStat(nil, info.FullMethod, "", costMs)
		if statErr != nil {
			remotelogs.Error("API_NODE", "create method stat failed: "+statErr.Error())
		}

		var tagMap = traceCtx.TagMap()
		for tag, tagCostMs := range tagMap {
			statErr = models.SharedAPIMethodStatDAO.CreateStat(nil, info.FullMethod, tag, tagCostMs)
			if statErr != nil {
				remotelogs.Error("API_NODE", "create method stat failed: "+statErr.Error())
			}
		}

		return
	}
	result, err := handler(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			err = status.Error(statusErr.Code(), "'"+info.FullMethod+"()' says: "+err.Error())
		} else {
			err = errors.New("'" + info.FullMethod + "()' says: " + err.Error())
		}
	}
	return result, err
}

// 添加启动相关的Issue
func (this *APINode) addStartIssue(code string, message string, suggestion string) {
	this.issues = append(this.issues, NewStartIssue(code, message, suggestion))
	this.saveIssues()
}

// 增加数据库建议
func (this *APINode) dbIssueSuggestion(errString string) string {
	// 数据库配置
	db, err := dbs.Default()
	if err != nil {
		return ""
	}
	config, err := db.Config()
	if err != nil {
		return ""
	}

	var dsn = config.Dsn
	dsnConfig, err := mysql.ParseDSN(dsn)
	if err != nil {
		return ""
	}
	var addr = dsnConfig.Addr

	// 配置文件位置
	var dbConfigPath = Tea.ConfigFile("db.yaml")

	// 连接被拒绝
	if strings.Contains(errString, "connection refused") {
		// 本机
		if strings.HasPrefix(addr, "127.0.0.1:") || strings.HasPrefix(addr, "localhost:") {
			return "试图连接到数据库被拒绝，请检查：1）本地数据库服务是否已经启动；2）数据库IP和端口（" + addr + "）是否正确；（当前数据库配置为：" + dsn + "，配置文件位置：" + dbConfigPath + "）。"
		} else {
			return "试图连接到数据库被拒绝，请检查：1）数据库服务是否已经启动；2）数据库IP和端口（" + addr + "）是否正确；3）防火墙设置；（当前数据库配置为：" + dsn + "，配置文件位置：" + dbConfigPath + "）。"
		}
	}

	// 权限错误
	if strings.Contains(errString, "Error 1045") || strings.Contains(errString, "Error 1044") {
		return "使用的用户和密码没有权限连接到指定数据库，请检查：1）数据库配置文件中的用户名（" + dsnConfig.User + "）和密码（" + dsnConfig.Passwd + "）是否正确；2）使用的用户是否已经在数据库中设置了正确的权限；（当前数据库配置为：" + dsn + "，配置文件位置：" + dbConfigPath + "）。"
	}

	// 数据库名称错误
	if strings.Contains(errString, "Error 1049") {
		return "数据库名称配置错误，请检查：数据库配置文件中数据库名称（" + dsnConfig.DBName + "）是否正确；（当前数据库配置为：" + dsn + "，配置文件位置：" + dbConfigPath + "）。"
	}

	return ""
}

// 保存issues
func (this *APINode) saveIssues() {
	issuesJSON, err := json.Marshal(this.issues)
	if err == nil {
		_ = os.WriteFile(this.issuesFile, issuesJSON, 0666)
	}
}

// 设置启动进度
func (this *APINode) setProgress(name, description string) {
	this.progress = &utils.Progress{
		Name:        name,
		Description: description,
	}
}

// 设置时区
func (this *APINode) setupTimeZone() {
	config, err := models.SharedSysSettingDAO.ReadAdminUIConfig(nil, nil)
	if err == nil && config != nil {
		if len(config.TimeZone) == 0 {
			config.TimeZone = nodeconfigs.DefaultTimeZoneLocation
		}
		location, err := time.LoadLocation(config.TimeZone)
		if err == nil && time.Local != location {
			time.Local = location
		}
	}
}
