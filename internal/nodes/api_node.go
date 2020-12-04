package nodes

import (
	"crypto/tls"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"net"
	"os"
	"strconv"
)

var sharedAPIConfig *configs.APIConfig = nil

type APINode struct {
}

func NewAPINode() *APINode {
	return &APINode{}
}

func (this *APINode) Start() {
	logs.Println("[API_NODE]start api node, pid: " + strconv.Itoa(os.Getpid()))

	// 自动升级
	err := this.autoUpgrade()
	if err != nil {
		logs.Println("[API_NODE]auto upgrade failed: " + err.Error())
		return
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
	apiNode, err := models.SharedAPINodeDAO.FindEnabledAPINodeWithUniqueIdAndSecret(config.NodeId, config.Secret)
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

	// 监听RPC服务
	logs.Println("[API_NODE]starting rpc ...")

	// HTTP
	httpConfig, err := apiNode.DecodeHTTP()
	if err != nil {
		logs.Println("[API_NODE]decode http config: " + err.Error())
		return
	}
	isListening := false
	if httpConfig != nil && httpConfig.IsOn && len(httpConfig.Listen) > 0 {
		for _, listen := range httpConfig.Listen {
			for _, addr := range listen.Addresses() {
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					logs.Println("[API_NODE]listening '" + addr + "' failed: " + err.Error())
					continue
				}
				go func() {
					err := this.listenRPC(listener, nil)
					if err != nil {
						logs.Println("[API_NODE]listening '" + addr + "' rpc: " + err.Error())
						return
					}
				}()
				isListening = true
			}
		}
	}

	// HTTPS
	httpsConfig, err := apiNode.DecodeHTTPS()
	if err != nil {
		logs.Println("[API_NODE]decode https config: " + err.Error())
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
					logs.Println("[API_NODE]listening '" + addr + "' failed: " + err.Error())
					continue
				}
				go func() {
					err := this.listenRPC(listener, &tls.Config{
						Certificates: certs,
					})
					if err != nil {
						logs.Println("[API_NODE]listening '" + addr + "' rpc: " + err.Error())
						return
					}
				}()
				isListening = true
			}
		}
	}

	if !isListening {
		logs.Println("[API_NODE]the api node does have a listening address")
		return
	}

	// 保持进程
	select {}
}

// 启动RPC监听
func (this *APINode) listenRPC(listener net.Listener, tlsConfig *tls.Config) error {
	var rpcServer *grpc.Server
	if tlsConfig == nil {
		logs.Println("[API_NODE]listening http://" + listener.Addr().String() + " ...")
		rpcServer = grpc.NewServer()
	} else {
		logs.Println("[API_NODE]listening https://" + listener.Addr().String() + " ...")
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
	pb.RegisterNodeGroupServiceServer(rpcServer, &services.NodeGroupService{})
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
	logs.Println("[API_NODE]upgrade database starting ...")
	err = setup.NewSQLExecutor(dbConfig).Run()
	if err != nil {
		return errors.New("execute sql failed: " + err.Error())
	}
	logs.Println("[API_NODE]upgrade database done")
	return nil
}
