package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

type ServerService struct {
}

// 创建服务
func (this *ServerService) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.CreateServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	serverId, err := models.SharedServerDAO.CreateServer(req.AdminId, req.UserId, req.Type, req.Name, req.Description, string(req.ServerNamesJON), string(req.HttpJSON), string(req.HttpsJSON), string(req.TcpJSON), string(req.TlsJSON), string(req.UnixJSON), string(req.UdpJSON), req.WebId, req.ReverseProxyJSON, req.ClusterId, string(req.IncludeNodesJSON), string(req.ExcludeNodesJSON))
	if err != nil {
		return nil, err
	}

	// 更新节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(req.ClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateServerResponse{ServerId: serverId}, nil
}

// 修改服务
func (this *ServerService) UpdateServerBasic(ctx context.Context, req *pb.UpdateServerBasicRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	err = models.SharedServerDAO.UpdateServerBasic(req.ServerId, req.Name, req.Description, req.ClusterId)
	if err != nil {
		return nil, err
	}

	// 更新老的节点版本
	if req.ClusterId != int64(server.ClusterId) {
		err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
		if err != nil {
			return nil, err
		}
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(req.ClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改HTTP服务
func (this *ServerService) UpdateServerHTTP(ctx context.Context, req *pb.UpdateServerHTTPRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerHTTP(req.ServerId, req.Config)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改HTTPS服务
func (this *ServerService) UpdateServerHTTPS(ctx context.Context, req *pb.UpdateServerHTTPSRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerHTTPS(req.ServerId, req.Config)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改TCP服务
func (this *ServerService) UpdateServerTCP(ctx context.Context, req *pb.UpdateServerTCPRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerTCP(req.ServerId, req.Config)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改TLS服务
func (this *ServerService) UpdateServerTLS(ctx context.Context, req *pb.UpdateServerTLSRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerTLS(req.ServerId, req.Config)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改Unix服务
func (this *ServerService) UpdateServerUnix(ctx context.Context, req *pb.UpdateServerUnixRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerUnix(req.ServerId, req.Config)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改UDP服务
func (this *ServerService) UpdateServerUDP(ctx context.Context, req *pb.UpdateServerUDPRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerUDP(req.ServerId, req.Config)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改Web服务
func (this *ServerService) UpdateServerWeb(ctx context.Context, req *pb.UpdateServerWebRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerWeb(req.ServerId, req.WebId)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改反向代理服务
func (this *ServerService) UpdateServerReverseProxy(ctx context.Context, req *pb.UpdateServerReverseProxyRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerReverseProxy(req.ServerId, req.ReverseProxyJSON)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改域名服务
func (this *ServerService) UpdateServerNames(ctx context.Context, req *pb.UpdateServerNamesRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerNames(req.ServerId, req.Config)
	if err != nil {
		return nil, err
	}

	// 更新新的节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 计算服务数量
func (this *ServerService) CountAllEnabledServers(ctx context.Context, req *pb.CountAllEnabledServersRequest) (*pb.CountAllEnabledServersResponse, error) {
	_ = req
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedServerDAO.CountAllEnabledServers()
	if err != nil {
		return nil, err
	}

	return &pb.CountAllEnabledServersResponse{Count: count}, nil
}

// 列出单页服务
func (this *ServerService) ListEnabledServers(ctx context.Context, req *pb.ListEnabledServersRequest) (*pb.ListEnabledServersResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	servers, err := models.SharedServerDAO.ListEnabledServers(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.Server{}
	for _, server := range servers {
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(server.ClusterId))
		if err != nil {
			return nil, err
		}
		result = append(result, &pb.Server{
			Id:           int64(server.Id),
			Type:         server.Type,
			Config:       []byte(server.Config),
			Name:         server.Name,
			Description:  server.Description,
			HttpJSON:     []byte(server.Http),
			HttpsJSON:    []byte(server.Https),
			TcpJSON:      []byte(server.Tcp),
			TlsJSON:      []byte(server.Tls),
			UnixJSON:     []byte(server.Unix),
			UdpJSON:      []byte(server.Udp),
			IncludeNodes: []byte(server.IncludeNodes),
			ExcludeNodes: []byte(server.ExcludeNodes),
			CreatedAt:    int64(server.CreatedAt),
			Cluster: &pb.NodeCluster{
				Id:   int64(server.ClusterId),
				Name: clusterName,
			},
		})
	}

	return &pb.ListEnabledServersResponse{Servers: result}, nil
}

// 禁用某服务
func (this *ServerService) DisableServer(ctx context.Context, req *pb.DisableServerRequest) (*pb.DisableServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	// 查找服务
	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find the server")
	}

	// 禁用服务
	err = models.SharedServerDAO.DisableServer(req.ServerId)
	if err != nil {
		return nil, err
	}

	// 更新节点版本
	err = models.SharedNodeDAO.UpdateAllNodesLatestVersionMatch(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.DisableServerResponse{}, nil
}

// 查找单个服务
func (this *ServerService) FindEnabledServer(ctx context.Context, req *pb.FindEnabledServerRequest) (*pb.FindEnabledServerResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	server, err := models.SharedServerDAO.FindEnabledServer(req.ServerId)
	if err != nil {
		return nil, err
	}

	if server == nil {
		return &pb.FindEnabledServerResponse{}, nil
	}

	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledServerResponse{Server: &pb.Server{
		Id:               int64(server.Id),
		Type:             server.Type,
		Name:             server.Name,
		Description:      server.Description,
		Config:           []byte(server.Config),
		ServerNamesJON:   []byte(server.ServerNames),
		HttpJSON:         []byte(server.Http),
		HttpsJSON:        []byte(server.Https),
		TcpJSON:          []byte(server.Tcp),
		TlsJSON:          []byte(server.Tls),
		UnixJSON:         []byte(server.Unix),
		UdpJSON:          []byte(server.Udp),
		WebId:            int64(server.WebId),
		ReverseProxyJSON: []byte(server.ReverseProxy),

		IncludeNodes: []byte(server.IncludeNodes),
		ExcludeNodes: []byte(server.ExcludeNodes),
		CreatedAt:    int64(server.CreatedAt),
		Cluster: &pb.NodeCluster{
			Id:   int64(server.ClusterId),
			Name: clusterName,
		},
	}}, nil
}

//
func (this *ServerService) FindEnabledServerType(ctx context.Context, req *pb.FindEnabledServerTypeRequest) (*pb.FindEnabledServerTypeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	serverType, err := models.SharedServerDAO.FindEnabledServerType(req.ServerId)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledServerTypeResponse{Type: serverType}, nil
}

// 查找反向代理设置
func (this *ServerService) FindAndInitServerReverseProxyConfig(ctx context.Context, req *pb.FindAndInitServerReverseProxyConfigRequest) (*pb.FindAndInitServerReverseProxyConfigResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	reverseProxyRef, err := models.SharedServerDAO.FindReverseProxyRef(req.ServerId)
	if err != nil {
		return nil, err
	}

	if reverseProxyRef == nil {
		reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(nil, nil, nil)
		if err != nil {
			return nil, err
		}

		reverseProxyRef = &serverconfigs.ReverseProxyRef{
			IsOn:           false,
			ReverseProxyId: reverseProxyId,
		}
		refJSON, err := json.Marshal(reverseProxyRef)
		if err != nil {
			return nil, err
		}
		err = models.SharedServerDAO.UpdateServerReverseProxy(req.ServerId, refJSON)
		if err != nil {
			return nil, err
		}
	}

	reverseProxyConfig, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(reverseProxyRef.ReverseProxyId)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(reverseProxyConfig)
	if err != nil {
		return nil, err
	}

	refJSON, err := json.Marshal(reverseProxyRef)
	if err != nil {
		return nil, err
	}

	return &pb.FindAndInitServerReverseProxyConfigResponse{ReverseProxyJSON: configJSON, ReverseProxyRefJSON: refJSON}, nil
}

// 初始化Web设置
func (this *ServerService) FindAndInitServerWebConfig(ctx context.Context, req *pb.FindAndInitServerWebConfigRequest) (*pb.FindAndInitServerWebConfigResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	webId, err := models.SharedServerDAO.FindServerWebId(req.ServerId)
	if err != nil {
		return nil, err
	}

	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(webId)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindAndInitServerWebConfigResponse{WebJSON: configJSON}, nil
}
