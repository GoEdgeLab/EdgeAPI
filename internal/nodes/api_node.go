package nodes

import (
	"crypto/tls"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	logs.Println("[API]start api node, pid: " + strconv.Itoa(os.Getpid()))

	// 读取配置
	config, err := configs.SharedAPIConfig()
	if err != nil {
		logs.Println("[API]start failed: " + err.Error())
		return
	}
	sharedAPIConfig = config

	// 校验
	apiNode, err := models.SharedAPINodeDAO.FindEnabledAPINodeWithUniqueIdAndSecret(config.NodeId, config.Secret)
	if err != nil {
		logs.Println("[API]start failed: read api node from database failed: " + err.Error())
		return
	}
	if apiNode == nil {
		logs.Println("[API]can not start node, wrong 'nodeId' or 'secret'")
		return
	}
	config.SetNumberId(int64(apiNode.Id))

	// 设置rlimit
	_ = utils.SetRLimit(1024 * 1024)

	// 监听RPC服务
	logs.Println("[API]starting rpc ...")

	// HTTP
	httpConfig, err := apiNode.DecodeHTTP()
	if err != nil {
		logs.Println("[API]decode http config: " + err.Error())
		return
	}
	isListening := false
	if httpConfig != nil && httpConfig.IsOn && len(httpConfig.Listen) > 0 {
		for _, listen := range httpConfig.Listen {
			for _, addr := range listen.Addresses() {
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					logs.Println("[API]listening '" + addr + "' failed: " + err.Error())
					continue
				}
				go func() {
					err := this.listenRPC(listener, nil)
					if err != nil {
						logs.Println("[API]listening '" + addr + "' rpc: " + err.Error())
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
		logs.Println("[API]decode https config: " + err.Error())
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
					logs.Println("[API]listening '" + addr + "' failed: " + err.Error())
					continue
				}
				go func() {
					err := this.listenRPC(listener, &tls.Config{
						Certificates: certs,
					})
					if err != nil {
						logs.Println("[API]listening '" + addr + "' rpc: " + err.Error())
						return
					}
				}()
				isListening = true
			}
		}
	}

	if !isListening {
		logs.Println("[API]the api node does have a listening address")
		return
	}

	// 保持进程
	select {}
}

// 启动RPC监听
func (this *APINode) listenRPC(listener net.Listener, tlsConfig *tls.Config) error {
	var rpcServer *grpc.Server
	if tlsConfig == nil {
		logs.Println("[API]listening http://" + listener.Addr().String() + " ...")
		rpcServer = grpc.NewServer()
	} else {
		logs.Println("[API]listening https://" + listener.Addr().String() + " ...")
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
	err := rpcServer.Serve(listener)
	if err != nil {
		return errors.New("[API]start rpc failed: " + err.Error())
	}

	return nil
}
