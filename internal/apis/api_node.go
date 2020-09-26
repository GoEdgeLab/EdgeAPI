package apis

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/logs"
	"google.golang.org/grpc"
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

	config, err := configs.SharedAPIConfig()
	if err != nil {
		logs.Println("[API]start failed: " + err.Error())
		return
	}
	sharedAPIConfig = config

	// 设置rlimit
	_ = utils.SetRLimit(1024 * 1024)

	// 监听RPC服务
	logs.Println("[API]start rpc: " + config.RPC.Listen)
	err = this.listenRPC()
	if err != nil {
		logs.Println(err.Error())
		return
	}
}

// 启动RPC监听
func (this *APINode) listenRPC() error {
	listener, err := net.Listen("tcp", sharedAPIConfig.RPC.Listen)
	if err != nil {
		return errors.New("[API]listen rpc failed: " + err.Error())
	}
	rpcServer := grpc.NewServer()
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
	err = rpcServer.Serve(listener)
	if err != nil {
		return errors.New("[API]start rpc failed: " + err.Error())
	}

	return nil
}
