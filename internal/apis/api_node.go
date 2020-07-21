package apis

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/log"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/monitor"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/node"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/provider"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/stat"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/user"
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
	dns.RegisterServiceServer(rpcServer, &dns.Service{})
	log.RegisterServiceServer(rpcServer, &log.Service{})
	monitor.RegisterServiceServer(rpcServer, &monitor.Service{})
	node.RegisterServiceServer(rpcServer, &node.Service{})
	provider.RegisterServiceServer(rpcServer, &provider.Service{})
	stat.RegisterServiceServer(rpcServer, &stat.Service{})
	user.RegisterServiceServer(rpcServer, &user.Service{})
	err = rpcServer.Serve(listener)
	if err != nil {
		return errors.New("[API]start rpc failed: " + err.Error())
	}

	return nil
}
