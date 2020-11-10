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
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	serverId, err := models.SharedServerDAO.CreateServer(req.AdminId, req.UserId, req.Type, req.Name, req.Description, string(req.ServerNamesJON), string(req.HttpJSON), string(req.HttpsJSON), string(req.TcpJSON), string(req.TlsJSON), string(req.UnixJSON), string(req.UdpJSON), req.WebId, req.ReverseProxyJSON, req.ClusterId, string(req.IncludeNodesJSON), string(req.ExcludeNodesJSON), req.GroupIds)
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
	// 校验请求
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

	err = models.SharedServerDAO.UpdateServerBasic(req.ServerId, req.Name, req.Description, req.ClusterId, req.IsOn, req.GroupIds)
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
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改HTTPS服务
func (this *ServerService) UpdateServerHTTPS(ctx context.Context, req *pb.UpdateServerHTTPSRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改TCP服务
func (this *ServerService) UpdateServerTCP(ctx context.Context, req *pb.UpdateServerTCPRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改TLS服务
func (this *ServerService) UpdateServerTLS(ctx context.Context, req *pb.UpdateServerTLSRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改Unix服务
func (this *ServerService) UpdateServerUnix(ctx context.Context, req *pb.UpdateServerUnixRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改UDP服务
func (this *ServerService) UpdateServerUDP(ctx context.Context, req *pb.UpdateServerUDPRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改Web服务
func (this *ServerService) UpdateServerWeb(ctx context.Context, req *pb.UpdateServerWebRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改反向代理服务
func (this *ServerService) UpdateServerReverseProxy(ctx context.Context, req *pb.UpdateServerReverseProxyRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改域名服务
func (this *ServerService) UpdateServerNames(ctx context.Context, req *pb.UpdateServerNamesRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
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

	return &pb.RPCUpdateSuccess{}, nil
}

// 计算服务数量
func (this *ServerService) CountAllEnabledServersMatch(ctx context.Context, req *pb.CountAllEnabledServersMatchRequest) (*pb.CountAllEnabledServersMatchResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedServerDAO.CountAllEnabledServersMatch(req.GroupId, req.Keyword)
	if err != nil {
		return nil, err
	}

	return &pb.CountAllEnabledServersMatchResponse{Count: count}, nil
}

// 列出单页服务
func (this *ServerService) ListEnabledServersMatch(ctx context.Context, req *pb.ListEnabledServersMatchRequest) (*pb.ListEnabledServersMatchResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	servers, err := models.SharedServerDAO.ListEnabledServersMatch(req.Offset, req.Size, req.GroupId, req.Keyword)
	if err != nil {
		return nil, err
	}
	result := []*pb.Server{}
	for _, server := range servers {
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(server.ClusterId))
		if err != nil {
			return nil, err
		}

		// 分组信息
		pbGroups := []*pb.ServerGroup{}
		if len(server.GroupIds) > 0 {
			groupIds := []int64{}
			err = json.Unmarshal([]byte(server.GroupIds), &groupIds)
			if err != nil {
				return nil, err
			}
			for _, groupId := range groupIds {
				group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(groupId)
				if err != nil {
					return nil, err
				}
				if group == nil {
					continue
				}
				pbGroups = append(pbGroups, &pb.ServerGroup{
					Id:   int64(group.Id),
					Name: group.Name,
				})
			}
		}

		result = append(result, &pb.Server{
			Id:             int64(server.Id),
			IsOn:           server.IsOn == 1,
			Type:           server.Type,
			Config:         []byte(server.Config),
			Name:           server.Name,
			Description:    server.Description,
			HttpJSON:       []byte(server.Http),
			HttpsJSON:      []byte(server.Https),
			TcpJSON:        []byte(server.Tcp),
			TlsJSON:        []byte(server.Tls),
			UnixJSON:       []byte(server.Unix),
			UdpJSON:        []byte(server.Udp),
			IncludeNodes:   []byte(server.IncludeNodes),
			ExcludeNodes:   []byte(server.ExcludeNodes),
			ServerNamesJON: []byte(server.ServerNames),
			CreatedAt:      int64(server.CreatedAt),
			Cluster: &pb.NodeCluster{
				Id:   int64(server.ClusterId),
				Name: clusterName,
			},
			Groups: pbGroups,
		})
	}

	return &pb.ListEnabledServersMatchResponse{Servers: result}, nil
}

// 禁用某服务
func (this *ServerService) DisableServer(ctx context.Context, req *pb.DisableServerRequest) (*pb.DisableServerResponse, error) {
	// 校验请求
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
	// 校验请求
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

	// 集群信息
	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	// 分组信息
	pbGroups := []*pb.ServerGroup{}
	if len(server.GroupIds) > 0 {
		groupIds := []int64{}
		err = json.Unmarshal([]byte(server.GroupIds), &groupIds)
		if err != nil {
			return nil, err
		}
		for _, groupId := range groupIds {
			group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(groupId)
			if err != nil {
				return nil, err
			}
			if group == nil {
				continue
			}
			pbGroups = append(pbGroups, &pb.ServerGroup{
				Id:   int64(group.Id),
				Name: group.Name,
			})
		}
	}

	return &pb.FindEnabledServerResponse{Server: &pb.Server{
		Id:               int64(server.Id),
		IsOn:             server.IsOn == 1,
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
		Groups: pbGroups,
	}}, nil
}

//
func (this *ServerService) FindEnabledServerType(ctx context.Context, req *pb.FindEnabledServerTypeRequest) (*pb.FindEnabledServerTypeResponse, error) {
	// 校验请求
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
	// 校验请求
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
	// 校验请求
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

// 计算使用某个SSL证书的服务数量
func (this *ServerService) CountAllEnabledServersWithSSLCertId(ctx context.Context, req *pb.CountAllEnabledServersWithSSLCertIdRequest) (*pb.CountAllEnabledServersWithSSLCertIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(req.CertId)
	if err != nil {
		return nil, err
	}

	if len(policyIds) == 0 {
		return &pb.CountAllEnabledServersWithSSLCertIdResponse{Count: 0}, nil
	}

	count, err := models.SharedServerDAO.CountAllEnabledServersWithSSLPolicyIds(policyIds)
	if err != nil {
		return nil, err
	}

	return &pb.CountAllEnabledServersWithSSLCertIdResponse{Count: count}, nil
}

// 查找使用某个SSL证书的所有服务
func (this *ServerService) FindAllEnabledServersWithSSLCertId(ctx context.Context, req *pb.FindAllEnabledServersWithSSLCertIdRequest) (*pb.FindAllEnabledServersWithSSLCertIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(req.CertId)
	if err != nil {
		return nil, err
	}
	if len(policyIds) == 0 {
		return &pb.FindAllEnabledServersWithSSLCertIdResponse{Servers: nil}, nil
	}

	servers, err := models.SharedServerDAO.FindAllEnabledServersWithSSLPolicyIds(policyIds)
	if err != nil {
		return nil, err
	}
	result := []*pb.Server{}
	for _, server := range servers {
		result = append(result, &pb.Server{
			Id:   int64(server.Id),
			Name: server.Name,
			IsOn: server.IsOn == 1,
			Type: server.Type,
		})
	}
	return &pb.FindAllEnabledServersWithSSLCertIdResponse{Servers: result}, nil
}

// 计算使用某个缓存策略的服务数量
func (this *ServerService) CountAllEnabledServersWithCachePolicyId(ctx context.Context, req *pb.CountAllEnabledServersWithCachePolicyIdRequest) (*pb.CountAllEnabledServersWithCachePolicyIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	webIds, err := models.SharedHTTPWebDAO.FindAllWebIdsWithCachePolicyId(req.CachePolicyId)
	if err != nil {
		return nil, err
	}
	if len(webIds) == 0 {
		return &pb.CountAllEnabledServersWithCachePolicyIdResponse{Count: 0}, nil
	}

	countServers, err := models.SharedServerDAO.CountEnabledServersWithWebIds(webIds)
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledServersWithCachePolicyIdResponse{Count: countServers}, nil
}

// 查找使用某个缓存策略的所有服务
func (this *ServerService) FindAllEnabledServersWithCachePolicyId(ctx context.Context, req *pb.FindAllEnabledServersWithCachePolicyIdRequest) (*pb.FindAllEnabledServersWithCachePolicyIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	webIds, err := models.SharedHTTPWebDAO.FindAllWebIdsWithCachePolicyId(req.CachePolicyId)
	if err != nil {
		return nil, err
	}

	if len(webIds) == 0 {
		return &pb.FindAllEnabledServersWithCachePolicyIdResponse{Servers: nil}, nil
	}

	servers, err := models.SharedServerDAO.FindAllEnabledServersWithWebIds(webIds)
	result := []*pb.Server{}
	for _, server := range servers {
		result = append(result, &pb.Server{
			Id:   int64(server.Id),
			Name: server.Name,
			IsOn: server.IsOn == 1,
			Type: server.Type,
			Cluster: &pb.NodeCluster{
				Id: int64(server.ClusterId),
			},
		})
	}
	return &pb.FindAllEnabledServersWithCachePolicyIdResponse{Servers: result}, nil
}

// 计算使用某个WAF策略的服务数量
func (this *ServerService) CountAllEnabledServersWithHTTPFirewallPolicyId(ctx context.Context, req *pb.CountAllEnabledServersWithHTTPFirewallPolicyIdRequest) (*pb.CountAllEnabledServersWithHTTPFirewallPolicyIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	webIds, err := models.SharedHTTPWebDAO.FindAllWebIdsWithHTTPFirewallPolicyId(req.FirewallPolicyId)
	if err != nil {
		return nil, err
	}

	if len(webIds) == 0 {
		return &pb.CountAllEnabledServersWithHTTPFirewallPolicyIdResponse{Count: 0}, nil
	}

	countServers, err := models.SharedServerDAO.CountEnabledServersWithWebIds(webIds)
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledServersWithHTTPFirewallPolicyIdResponse{Count: countServers}, nil
}

// 查找使用某个WAF策略的所有服务
func (this *ServerService) FindAllEnabledServersWithHTTPFirewallPolicyId(ctx context.Context, req *pb.FindAllEnabledServersWithHTTPFirewallPolicyIdRequest) (*pb.FindAllEnabledServersWithHTTPFirewallPolicyIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	webIds, err := models.SharedHTTPWebDAO.FindAllWebIdsWithHTTPFirewallPolicyId(req.FirewallPolicyId)
	if err != nil {
		return nil, err
	}

	if len(webIds) == 0 {
		return &pb.FindAllEnabledServersWithHTTPFirewallPolicyIdResponse{Servers: nil}, nil
	}

	servers, err := models.SharedServerDAO.FindAllEnabledServersWithWebIds(webIds)
	result := []*pb.Server{}
	for _, server := range servers {
		result = append(result, &pb.Server{
			Id:   int64(server.Id),
			Name: server.Name,
			IsOn: server.IsOn == 1,
			Type: server.Type,
			Cluster: &pb.NodeCluster{
				Id: int64(server.ClusterId),
			},
		})
	}

	return &pb.FindAllEnabledServersWithHTTPFirewallPolicyIdResponse{Servers: result}, nil
}

// 计算运行在某个集群上的所有服务数量
func (this *ServerService) CountAllEnabledServersWithNodeClusterId(ctx context.Context, req *pb.CountAllEnabledServersWithNodeClusterIdRequest) (*pb.CountAllEnabledServersWithNodeClusterIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedServerDAO.CountAllEnabledServersWithNodeClusterId(req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledServersWithNodeClusterIdResponse{Count: count}, nil
}

// 计算使用某个分组的服务数量
func (this *ServerService) CountAllEnabledServersWithGroupId(ctx context.Context, req *pb.CountAllEnabledServersWithGroupIdRequest) (*pb.CountAllEnabledServersWithGroupIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedServerDAO.CountAllEnabledServersWithGroupId(req.GroupId)
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledServersWithGroupIdResponse{
		Count: count,
	}, nil
}

// 通知更新
func (this *ServerService) NotifyServersChange(ctx context.Context, req *pb.NotifyServersChangeRequest) (*pb.NotifyServersChangeResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedSysEventDAO.CreateEvent(models.NewServerChangeEvent())
	if err != nil {
		return nil, err
	}

	return &pb.NotifyServersChangeResponse{}, nil
}
