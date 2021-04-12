package nodes

import (
	"crypto/tls"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/events"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var sharedAPIConfig *configs.APIConfig = nil

type APINode struct {
}

func NewAPINode() *APINode {
	return &APINode{}
}

func (this *APINode) Start() {
	logs.Println("[API_NODE]start api node, pid: " + strconv.Itoa(os.Getpid()))

	// 本地Sock
	err := this.listenSock()
	if err != nil {
		logs.Println("[API_NODE]" + err.Error())
		return
	}

	// 自动升级
	err = this.autoUpgrade()
	if err != nil {
		logs.Println("[API_NODE]auto upgrade failed: " + err.Error())
		return
	}

	// 自动设置数据库
	err = this.setupDB()
	if err != nil {
		logs.Println("[API_NODE]setup database '" + err.Error() + "'")

		// 不阻断执行
	}

	// 数据库通知启动
	dbs.NotifyReady()

	// 读取配置
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
	go NewNodeStatusExecutor().Listen()

	// 监听RPC服务
	remotelogs.Println("API_NODE", "starting RPC server ...")

	isListening := this.listenPorts(apiNode)

	if !isListening {
		remotelogs.Error("API_NODE", "the api node require at least one listening address")
		return
	}

	// 保持进程
	select {}
}

// 实现守护进程
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

// 安装系统服务
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
		rpcServer = grpc.NewServer()
	} else {
		logs.Println("[API_NODE]listening GRPC https://" + listener.Addr().String() + " ...")
		rpcServer = grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	pb.RegisterAdminServiceServer(rpcServer, &services.AdminService{})
	pb.RegisterNodeGrantServiceServer(rpcServer, &services.NodeGrantService{})
	pb.RegisterServerServiceServer(rpcServer, &services.ServerService{})
	pb.RegisterNodeServiceServer(rpcServer, &services.NodeService{})
	pb.RegisterNodeClusterServiceServer(rpcServer, &services.NodeClusterService{})
	pb.RegisterNodeIPAddressServiceServer(rpcServer, &services.NodeIPAddressService{})
	pb.RegisterAPINodeServiceServer(rpcServer, &services.APINodeService{})
	pb.RegisterOriginServiceServer(rpcServer, &services.OriginService{})
	pb.RegisterHTTPWebServiceServer(rpcServer, &services.HTTPWebService{})
	pb.RegisterReverseProxyServiceServer(rpcServer, &services.ReverseProxyService{})
	pb.RegisterHTTPGzipServiceServer(rpcServer, &services.HTTPGzipService{})
	pb.RegisterHTTPHeaderPolicyServiceServer(rpcServer, &services.HTTPHeaderPolicyService{})
	pb.RegisterHTTPHeaderServiceServer(rpcServer, &services.HTTPHeaderService{})
	pb.RegisterHTTPPageServiceServer(rpcServer, &services.HTTPPageService{})
	pb.RegisterHTTPAccessLogPolicyServiceServer(rpcServer, &services.HTTPAccessLogPolicyService{})
	pb.RegisterHTTPCachePolicyServiceServer(rpcServer, &services.HTTPCachePolicyService{})
	pb.RegisterHTTPFirewallPolicyServiceServer(rpcServer, &services.HTTPFirewallPolicyService{})
	pb.RegisterHTTPLocationServiceServer(rpcServer, &services.HTTPLocationService{})
	pb.RegisterHTTPWebsocketServiceServer(rpcServer, &services.HTTPWebsocketService{})
	pb.RegisterHTTPRewriteRuleServiceServer(rpcServer, &services.HTTPRewriteRuleService{})
	pb.RegisterSSLCertServiceServer(rpcServer, &services.SSLCertService{})
	pb.RegisterSSLPolicyServiceServer(rpcServer, &services.SSLPolicyService{})
	pb.RegisterSysSettingServiceServer(rpcServer, &services.SysSettingService{})
	pb.RegisterHTTPFirewallRuleGroupServiceServer(rpcServer, &services.HTTPFirewallRuleGroupService{})
	pb.RegisterHTTPFirewallRuleSetServiceServer(rpcServer, &services.HTTPFirewallRuleSetService{})
	pb.RegisterDBNodeServiceServer(rpcServer, &services.DBNodeService{})
	pb.RegisterNodeLogServiceServer(rpcServer, &services.NodeLogService{})
	pb.RegisterHTTPAccessLogServiceServer(rpcServer, &services.HTTPAccessLogService{})
	pb.RegisterMessageServiceServer(rpcServer, &services.MessageService{})
	pb.RegisterMessageRecipientServiceServer(rpcServer, &services.MessageRecipientService{})
	pb.RegisterMessageReceiverServiceServer(rpcServer, &services.MessageReceiverService{})
	pb.RegisterMessageMediaServiceServer(rpcServer, &services.MessageMediaService{})
	pb.RegisterMessageRecipientGroupServiceServer(rpcServer, &services.MessageRecipientGroupService{})
	pb.RegisterMessageMediaInstanceServiceServer(rpcServer, &services.MessageMediaInstanceService{})
	pb.RegisterMessageTaskServiceServer(rpcServer, &services.MessageTaskService{})
	pb.RegisterMessageTaskLogServiceServer(rpcServer, &services.MessageTaskLogService{})
	pb.RegisterNodeGroupServiceServer(rpcServer, &services.NodeGroupService{})
	pb.RegisterNodeRegionServiceServer(rpcServer, &services.NodeRegionService{})
	pb.RegisterNodePriceItemServiceServer(rpcServer, &services.NodePriceItemService{})
	pb.RegisterServerGroupServiceServer(rpcServer, &services.ServerGroupService{})
	pb.RegisterIPLibraryServiceServer(rpcServer, &services.IPLibraryService{})
	pb.RegisterFileChunkServiceServer(rpcServer, &services.FileChunkService{})
	pb.RegisterFileServiceServer(rpcServer, &services.FileService{})
	pb.RegisterRegionCountryServiceServer(rpcServer, &services.RegionCountryService{})
	pb.RegisterRegionProvinceServiceServer(rpcServer, &services.RegionProvinceService{})
	pb.RegisterIPListServiceServer(rpcServer, &services.IPListService{})
	pb.RegisterIPItemServiceServer(rpcServer, &services.IPItemService{})
	pb.RegisterLogServiceServer(rpcServer, &services.LogService{})
	pb.RegisterDNSProviderServiceServer(rpcServer, &services.DNSProviderService{})
	pb.RegisterDNSDomainServiceServer(rpcServer, &services.DNSDomainService{})
	pb.RegisterDNSServiceServer(rpcServer, &services.DNSService{})
	pb.RegisterACMEUserServiceServer(rpcServer, &services.ACMEUserService{})
	pb.RegisterACMETaskServiceServer(rpcServer, &services.ACMETaskService{})
	pb.RegisterACMEAuthenticationServiceServer(rpcServer, &services.ACMEAuthenticationService{})
	pb.RegisterUserServiceServer(rpcServer, &services.UserService{})
	pb.RegisterServerDailyStatServiceServer(rpcServer, &services.ServerDailyStatService{})
	pb.RegisterUserBillServiceServer(rpcServer, &services.UserBillService{})
	pb.RegisterUserNodeServiceServer(rpcServer, &services.UserNodeService{})
	pb.RegisterLoginServiceServer(rpcServer, &services.LoginService{})
	pb.RegisterUserAccessKeyServiceServer(rpcServer, &services.UserAccessKeyService{})
	pb.RegisterSysLockerServiceServer(rpcServer, &services.SysLockerService{})
	pb.RegisterNodeTaskServiceServer(rpcServer, &services.NodeTaskService{})
	pb.RegisterDBServiceServer(rpcServer, &services.DBService{})
	pb.RegisterServerRegionCityMonthlyStatServiceServer(rpcServer, &services.ServerRegionCityMonthlyStatService{})
	pb.RegisterServerRegionCountryMonthlyStatServiceServer(rpcServer, &services.ServerRegionCountryMonthlyStatService{})
	pb.RegisterServerRegionProvinceMonthlyStatServiceServer(rpcServer, &services.ServerRegionProvinceMonthlyStatService{})
	pb.RegisterServerRegionProviderMonthlyStatServiceServer(rpcServer, &services.ServerRegionProviderMonthlyStatService{})
	pb.RegisterServerClientSystemMonthlyStatServiceServer(rpcServer, &services.ServerClientSystemMonthlyStatService{})
	pb.RegisterServerClientBrowserMonthlyStatServiceServer(rpcServer, &services.ServerClientBrowserMonthlyStatService{})
	pb.RegisterServerHTTPFirewallDailyStatServiceServer(rpcServer, &services.ServerHTTPFirewallDailyStatService{})
	pb.RegisterDNSTaskServiceServer(rpcServer, &services.DNSTaskService{})
	pb.RegisterNodeClusterFirewallActionServiceServer(rpcServer, &services.NodeClusterFirewallActionService{})
	pb.RegisterMonitorNodeServiceServer(rpcServer, &services.MonitorNodeService{})
	err := rpcServer.Serve(listener)
	if err != nil {
		return errors.New("[API_NODE]start rpc failed: " + err.Error())
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
	// 不使用remotelogs()，因为此时还没有启动完成
	logs.Println("[API_NODE]upgrade database starting ...")
	err = setup.NewSQLExecutor(dbConfig).Run()
	if err != nil {
		return errors.New("execute sql failed: " + err.Error())
	}
	// 不使用remotelogs
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
					return err
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
					remotelogs.Error("API_NODE", "listening '"+addr+"' failed: "+err.Error())
					continue
				}
				go func() {
					err := this.listenRPC(listener, nil)
					if err != nil {
						remotelogs.Error("API_NODE", "listening '"+addr+"' rpc: "+err.Error())
						return
					}
				}()
				isListening = true
			}
		}
	}

	// HTTPS
	httpsConfig, err := apiNode.DecodeHTTPS(nil)
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
					remotelogs.Error("API_NODE", "listening '"+addr+"' failed: "+err.Error())
					continue
				}
				go func() {
					err := this.listenRPC(listener, &tls.Config{
						Certificates: certs,
					})
					if err != nil {
						remotelogs.Error("API_NODE", "listening '"+addr+"' rpc: "+err.Error())
						return
					}
				}()
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
				go func() {
					remotelogs.Println("API_NODE", "listening REST http://"+addr+" ...")
					server := &RestServer{}
					err := server.Listen(listener)
					if err != nil {
						remotelogs.Error("API_NODE", "listening REST 'http://"+addr+"' failed: "+err.Error())
						return
					}
				}()
				isListening = true
			}
		}
	}

	// Rest HTTPS
	restHTTPSConfig, err := apiNode.DecodeRestHTTPS(nil)
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
				go func() {
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
				}()
				isListening = true
			}
		}
	}

	return
}

// 监听本地sock
func (this *APINode) listenSock() error {
	path := os.TempDir() + "/edge-api.sock"

	// 检查是否已经存在
	_, err := os.Stat(path)
	if err == nil {
		conn, err := net.Dial("unix", path)
		if err != nil {
			_ = os.Remove(path)
		} else {
			_ = conn.Close()
		}
	}

	// 新的监听任务
	listener, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	events.On(events.EventQuit, func() {
		remotelogs.Println("API_NODE", "quit unix sock")
		_ = listener.Close()
	})

	go func() {
		for {
			_, err := listener.Accept()
			if err != nil {
				return
			}
		}
	}()

	return nil
}
