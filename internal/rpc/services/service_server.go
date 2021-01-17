package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

type ServerService struct {
	BaseService
}

// 创建服务
func (this *ServerService) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.CreateServerResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 校验用户相关数据
	if userId > 0 {
		// HTTPS
		if len(req.HttpsJSON) > 0 {
			httpsConfig := &serverconfigs.HTTPSProtocolConfig{}
			err = json.Unmarshal(req.HttpsJSON, httpsConfig)
			if err != nil {
				return nil, err
			}
			if httpsConfig.SSLPolicyRef != nil && httpsConfig.SSLPolicyRef.SSLPolicyId > 0 {
				err := models.SharedSSLPolicyDAO.CheckUserPolicy(tx, httpsConfig.SSLPolicyRef.SSLPolicyId, userId)
				if err != nil {
					return nil, err
				}
			}
		}

		// TLS
		if len(req.TlsJSON) > 0 {
			tlsConfig := &serverconfigs.TLSProtocolConfig{}
			err = json.Unmarshal(req.TlsJSON, tlsConfig)
			if err != nil {
				return nil, err
			}
			if tlsConfig.SSLPolicyRef != nil && tlsConfig.SSLPolicyRef.SSLPolicyId > 0 {
				err := models.SharedSSLPolicyDAO.CheckUserPolicy(tx, tlsConfig.SSLPolicyRef.SSLPolicyId, userId)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// 是否需要审核
	isAuditing := false
	serverNamesJSON := req.ServerNamesJON
	auditingServerNamesJSON := []byte("[]")
	if userId > 0 {
		// 如果域名不为空的时候需要审核
		if len(serverNamesJSON) > 0 && string(serverNamesJSON) != "[]" {
			globalConfig, err := models.SharedSysSettingDAO.ReadGlobalConfig(tx)
			if err != nil {
				return nil, err
			}
			if globalConfig != nil && globalConfig.HTTPAll.DomainAuditingIsOn {
				isAuditing = true
				serverNamesJSON = []byte("[]")
				auditingServerNamesJSON = req.ServerNamesJON
			}
		}
	}

	serverId, err := models.SharedServerDAO.CreateServer(tx, req.AdminId, req.UserId, req.Type, req.Name, req.Description, serverNamesJSON, isAuditing, auditingServerNamesJSON, string(req.HttpJSON), string(req.HttpsJSON), string(req.TcpJSON), string(req.TlsJSON), string(req.UnixJSON), string(req.UdpJSON), req.WebId, req.ReverseProxyJSON, req.NodeClusterId, string(req.IncludeNodesJSON), string(req.ExcludeNodesJSON), req.GroupIds)
	if err != nil {
		return nil, err
	}

	return &pb.CreateServerResponse{ServerId: serverId}, nil
}

// 修改服务基本信息
func (this *ServerService) UpdateServerBasic(ctx context.Context, req *pb.UpdateServerBasicRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	err = models.SharedServerDAO.UpdateServerBasic(tx, req.ServerId, req.Name, req.Description, req.NodeClusterId, req.IsOn, req.GroupIds)
	if err != nil {
		return nil, err
	}

	// 检查服务变化
	oldIsOn := server.IsOn == 1
	if oldIsOn != req.IsOn {
		go func() {
			err := this.notifyServerDNSChanged(req.ServerId)
			if err != nil {
				logs.Println("[DNS]notify server changed: " + err.Error())
			}
		}()
	}

	return this.Success()
}

// 修改服务是否启用
func (this *ServerService) UpdateServerIsOn(ctx context.Context, req *pb.UpdateServerIsOnRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}
	err = models.SharedServerDAO.UpdateServerIsOn(tx, req.ServerId, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 修改HTTP服务
func (this *ServerService) UpdateServerHTTP(ctx context.Context, req *pb.UpdateServerHTTPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerHTTP(tx, req.ServerId, req.HttpJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改HTTPS服务
func (this *ServerService) UpdateServerHTTPS(ctx context.Context, req *pb.UpdateServerHTTPSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerHTTPS(tx, req.ServerId, req.HttpsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改TCP服务
func (this *ServerService) UpdateServerTCP(ctx context.Context, req *pb.UpdateServerTCPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerTCP(tx, req.ServerId, req.TcpJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改TLS服务
func (this *ServerService) UpdateServerTLS(ctx context.Context, req *pb.UpdateServerTLSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerTLS(tx, req.ServerId, req.TlsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改Unix服务
func (this *ServerService) UpdateServerUnix(ctx context.Context, req *pb.UpdateServerUnixRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerUnix(tx, req.ServerId, req.UnixJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改UDP服务
func (this *ServerService) UpdateServerUDP(ctx context.Context, req *pb.UpdateServerUDPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerUDP(tx, req.ServerId, req.UdpJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改Web服务
func (this *ServerService) UpdateServerWeb(ctx context.Context, req *pb.UpdateServerWebRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerWeb(tx, req.ServerId, req.WebId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改反向代理服务
func (this *ServerService) UpdateServerReverseProxy(ctx context.Context, req *pb.UpdateServerReverseProxyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerReverseProxy(tx, req.ServerId, req.ReverseProxyJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找服务的域名设置
func (this *ServerService) FindServerNames(ctx context.Context, req *pb.FindServerNamesRequest) (*pb.FindServerNamesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	serverNamesJSON, isAuditing, auditingServerNamesJSON, auditingResultJSON, err := models.SharedServerDAO.FindServerNames(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	// 审核结果
	auditingResult := &pb.ServerNameAuditingResult{}
	if len(auditingResultJSON) > 0 {
		err = json.Unmarshal(auditingResultJSON, auditingResult)
		if err != nil {
			return nil, err
		}
	} else {
		auditingResult.IsOk = true
	}

	return &pb.FindServerNamesResponse{
		ServerNamesJSON:         serverNamesJSON,
		IsAuditing:              isAuditing,
		AuditingServerNamesJSON: auditingServerNamesJSON,
		AuditingResult:          auditingResult,
	}, nil
}

// 修改域名服务
func (this *ServerService) UpdateServerNames(ctx context.Context, req *pb.UpdateServerNamesRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	tx := this.NullTx()

	// 是否需要审核
	if userId > 0 {
		globalConfig, err := models.SharedSysSettingDAO.ReadGlobalConfig(tx)
		if err != nil {
			return nil, err
		}
		if globalConfig != nil && globalConfig.HTTPAll.DomainAuditingIsOn {
			err = models.SharedServerDAO.UpdateAuditingServerNames(tx, req.ServerId, true, req.ServerNamesJSON)
			if err != nil {
				return nil, err
			}
			return this.Success()
		}
	}

	// 修改配置
	err = models.SharedServerDAO.UpdateServerNames(tx, req.ServerId, req.ServerNamesJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 审核服务的域名设置
func (this *ServerService) UpdateServerNamesAuditing(ctx context.Context, req *pb.UpdateServerNamesAuditingRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if req.AuditingResult == nil {
		return nil, errors.New("'result' should not be nil")
	}

	tx := this.NullTx()

	err = models.SharedServerDAO.UpdateServerAuditing(tx, req.ServerId, req.AuditingResult)
	if err != nil {
		return nil, err
	}

	// 发送消息提醒
	_, userId, err := models.SharedServerDAO.FindServerAdminIdAndUserId(tx, req.ServerId)
	if userId > 0 {
		if req.AuditingResult.IsOk {
			err = models.SharedMessageDAO.CreateMessage(tx, 0, userId, models.MessageTypeServerNamesAuditingSuccess, models.LevelInfo, "服务域名审核通过", maps.Map{
				"serverId": req.ServerId,
			}.AsJSON())
			if err != nil {
				return nil, err
			}
		} else {
			err = models.SharedMessageDAO.CreateMessage(tx, 0, userId, models.MessageTypeServerNamesAuditingFailed, models.LevelError, "服务域名审核失败，原因："+req.AuditingResult.Reason, maps.Map{
				"serverId": req.ServerId,
			}.AsJSON())
			if err != nil {
				return nil, err
			}
		}
	}

	// 通知服务更新
	go func() {
		err := this.notifyServerDNSChanged(req.ServerId)
		if err != nil {
			logs.Println("[DNS]notify server changed: " + err.Error())
		}
	}()

	return this.Success()
}

// 计算服务数量
func (this *ServerService) CountAllEnabledServersMatch(ctx context.Context, req *pb.CountAllEnabledServersMatchRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersMatch(tx, req.GroupId, req.Keyword, req.UserId, req.ClusterId, types.Int8(req.AuditingFlag), req.ProtocolFamily)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// 列出单页服务
func (this *ServerService) ListEnabledServersMatch(ctx context.Context, req *pb.ListEnabledServersMatchRequest) (*pb.ListEnabledServersMatchResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	servers, err := models.SharedServerDAO.ListEnabledServersMatch(tx, req.Offset, req.Size, req.GroupId, req.Keyword, req.UserId, req.ClusterId, req.AuditingFlag, req.ProtocolFamily)
	if err != nil {
		return nil, err
	}
	result := []*pb.Server{}
	for _, server := range servers {
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(server.ClusterId))
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
				group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(tx, groupId)
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

		// 用户
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(server.UserId))
		if err != nil {
			return nil, err
		}
		var pbUser *pb.User = nil
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Fullname: user.Fullname,
			}
		}

		// 审核结果
		auditingResult := &pb.ServerNameAuditingResult{}
		if len(server.AuditingResult) > 0 {
			err = json.Unmarshal([]byte(server.AuditingResult), auditingResult)
			if err != nil {
				return nil, err
			}
		} else {
			auditingResult.IsOk = true
		}

		result = append(result, &pb.Server{
			Id:                      int64(server.Id),
			IsOn:                    server.IsOn == 1,
			Type:                    server.Type,
			Config:                  []byte(server.Config),
			Name:                    server.Name,
			Description:             server.Description,
			HttpJSON:                []byte(server.Http),
			HttpsJSON:               []byte(server.Https),
			TcpJSON:                 []byte(server.Tcp),
			TlsJSON:                 []byte(server.Tls),
			UnixJSON:                []byte(server.Unix),
			UdpJSON:                 []byte(server.Udp),
			IncludeNodes:            []byte(server.IncludeNodes),
			ExcludeNodes:            []byte(server.ExcludeNodes),
			ServerNamesJSON:         []byte(server.ServerNames),
			IsAuditing:              server.IsAuditing == 1,
			AuditingServerNamesJSON: []byte(server.AuditingServerNames),
			AuditingResult:          auditingResult,
			CreatedAt:               int64(server.CreatedAt),
			DnsName:                 server.DnsName,
			NodeCluster: &pb.NodeCluster{
				Id:   int64(server.ClusterId),
				Name: clusterName,
			},
			Groups: pbGroups,
			User:   pbUser,
		})
	}

	return &pb.ListEnabledServersMatchResponse{Servers: result}, nil
}

// 禁用某服务
func (this *ServerService) DeleteServer(ctx context.Context, req *pb.DeleteServerRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 禁用服务
	err = models.SharedServerDAO.DisableServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找单个服务
func (this *ServerService) FindEnabledServer(ctx context.Context, req *pb.FindEnabledServerRequest) (*pb.FindEnabledServerResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if server == nil {
		return &pb.FindEnabledServerResponse{}, nil
	}

	// 集群信息
	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(server.ClusterId))
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
			group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(tx, groupId)
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

	// 用户信息
	var pbUser *pb.User = nil
	if server.UserId > 0 {
		user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(server.UserId))
		if err != nil {
			return nil, err
		}
		if user != nil {
			pbUser = &pb.User{
				Id:       int64(user.Id),
				Username: user.Username,
				Fullname: user.Fullname,
			}
		}
	}

	return &pb.FindEnabledServerResponse{Server: &pb.Server{
		Id:               int64(server.Id),
		IsOn:             server.IsOn == 1,
		Type:             server.Type,
		Name:             server.Name,
		Description:      server.Description,
		DnsName:          server.DnsName,
		Config:           []byte(server.Config),
		ServerNamesJSON:  []byte(server.ServerNames),
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
		NodeCluster: &pb.NodeCluster{
			Id:   int64(server.ClusterId),
			Name: clusterName,
		},
		Groups: pbGroups,
		User:   pbUser,
	}}, nil
}

// 查找服务配置
func (this *ServerService) FindEnabledServerConfig(ctx context.Context, req *pb.FindEnabledServerConfigRequest) (*pb.FindEnabledServerConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedServerDAO.ComposeServerConfig(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return &pb.FindEnabledServerConfigResponse{ServerJSON: nil}, nil
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledServerConfigResponse{ServerJSON: configJSON}, nil
}

// 查找服务的服务类型
func (this *ServerService) FindEnabledServerType(ctx context.Context, req *pb.FindEnabledServerTypeRequest) (*pb.FindEnabledServerTypeResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	serverType, err := models.SharedServerDAO.FindEnabledServerType(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledServerTypeResponse{Type: serverType}, nil
}

// 查找反向代理设置
func (this *ServerService) FindAndInitServerReverseProxyConfig(ctx context.Context, req *pb.FindAndInitServerReverseProxyConfigRequest) (*pb.FindAndInitServerReverseProxyConfigResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	reverseProxyRef, err := models.SharedServerDAO.FindReverseProxyRef(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if reverseProxyRef == nil {
		reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, userId, nil, nil, nil)
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
		err = models.SharedServerDAO.UpdateServerReverseProxy(tx, req.ServerId, refJSON)
		if err != nil {
			return nil, err
		}
	}

	reverseProxyConfig, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	webId, err := models.SharedServerDAO.FindServerWebId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	if webId == 0 {
		webId, err = models.SharedServerDAO.InitServerWeb(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId)
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
func (this *ServerService) CountAllEnabledServersWithSSLCertId(ctx context.Context, req *pb.CountAllEnabledServersWithSSLCertIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}
	if userId > 0 {
		// TODO 校验权限
	}

	tx := this.NullTx()

	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}

	if len(policyIds) == 0 {
		return this.SuccessCount(0)
	}

	count, err := models.SharedServerDAO.CountAllEnabledServersWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// 查找使用某个SSL证书的所有服务
func (this *ServerService) FindAllEnabledServersWithSSLCertId(ctx context.Context, req *pb.FindAllEnabledServersWithSSLCertIdRequest) (*pb.FindAllEnabledServersWithSSLCertIdResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}

	tx := this.NullTx()

	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}
	if len(policyIds) == 0 {
		return &pb.FindAllEnabledServersWithSSLCertIdResponse{Servers: nil}, nil
	}

	servers, err := models.SharedServerDAO.FindAllEnabledServersWithSSLPolicyIds(tx, policyIds)
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

// 计算运行在某个集群上的所有服务数量
func (this *ServerService) CountAllEnabledServersWithNodeClusterId(ctx context.Context, req *pb.CountAllEnabledServersWithNodeClusterIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersWithNodeClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 计算使用某个分组的服务数量
func (this *ServerService) CountAllEnabledServersWithGroupId(ctx context.Context, req *pb.CountAllEnabledServersWithGroupIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersWithGroupId(tx, req.GroupId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 通知更新
func (this *ServerService) NotifyServersChange(ctx context.Context, req *pb.NotifyServersChangeRequest) (*pb.NotifyServersChangeResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	clusterIds, err := models.SharedNodeClusterDAO.FindAllEnableClusterIds(tx)
	if err != nil {
		return nil, err
	}
	for _, clusterId := range clusterIds {
		err = models.SharedNodeClusterDAO.NotifyUpdate(tx, clusterId)
		if err != nil {
			return nil, err
		}
	}

	return &pb.NotifyServersChangeResponse{}, nil
}

// 取得某个集群下的所有服务相关的DNS
func (this *ServerService) FindAllEnabledServersDNSWithClusterId(ctx context.Context, req *pb.FindAllEnabledServersDNSWithClusterIdRequest) (*pb.FindAllEnabledServersDNSWithClusterIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	servers, err := models.SharedServerDAO.FindAllServersDNSWithClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result := []*pb.ServerDNSInfo{}
	for _, server := range servers {
		// 如果子域名为空
		if len(server.DnsName) == 0 {
			// 自动生成子域名
			dnsName, err := models.SharedServerDAO.GenerateServerDNSName(tx, int64(server.Id))
			if err != nil {
				return nil, err
			}
			server.DnsName = dnsName
		}

		result = append(result, &pb.ServerDNSInfo{
			Id:      int64(server.Id),
			Name:    server.Name,
			DnsName: server.DnsName,
		})
	}

	return &pb.FindAllEnabledServersDNSWithClusterIdResponse{Servers: result}, nil
}

// 查找单个服务的DNS信息
func (this *ServerService) FindEnabledServerDNS(ctx context.Context, req *pb.FindEnabledServerDNSRequest) (*pb.FindEnabledServerDNSResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	dnsName, err := models.SharedServerDAO.FindServerDNSName(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	var pbDomain *pb.DNSDomain = nil
	if clusterId > 0 {
		clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId)
		if err != nil {
			return nil, err
		}
		if clusterDNS != nil {
			domainId := int64(clusterDNS.DnsDomainId)
			if domainId > 0 {
				domain, err := models.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId)
				if err != nil {
					return nil, err
				}
				if domain != nil {
					pbDomain = &pb.DNSDomain{
						Id:   domainId,
						Name: domain.Name,
					}
				}
			}
		}
	}

	return &pb.FindEnabledServerDNSResponse{
		DnsName: dnsName,
		Domain:  pbDomain,
	}, nil
}

// 自动同步DNS状态
func (this *ServerService) notifyServerDNSChanged(serverId int64) error {
	tx := this.NullTx()

	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, serverId)
	if err != nil {
		return err
	}
	dnsInfo, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId)
	if err != nil {
		return err
	}
	if dnsInfo == nil {
		return nil
	}
	if len(dnsInfo.DnsName) == 0 || dnsInfo.DnsDomainId == 0 {
		return nil
	}
	dnsConfig, err := dnsInfo.DecodeDNSConfig()
	if err != nil {
		return err
	}
	if !dnsConfig.ServersAutoSync {
		return nil
	}

	// 执行同步
	domainService := &DNSDomainService{}
	resp, err := domainService.syncClusterDNS(&pb.SyncDNSDomainDataRequest{
		DnsDomainId:   int64(dnsInfo.DnsDomainId),
		NodeClusterId: clusterId,
	})
	if err != nil {
		return err
	}
	if !resp.IsOk {
		err = models.SharedMessageDAO.CreateClusterMessage(tx, clusterId, models.MessageTypeClusterDNSSyncFailed, models.LevelError, "集群DNS同步失败："+resp.Error, nil)
		if err != nil {
			logs.Println("[NODE_SERVICE]" + err.Error())
		}
	}
	return nil
}

// 检查服务是否属于某个用户
func (this *ServerService) CheckUserServer(ctx context.Context, req *pb.CheckUserServerRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUser(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 查找一个用户下的所有域名列表
func (this *ServerService) FindAllEnabledServerNamesWithUserId(ctx context.Context, req *pb.FindAllEnabledServerNamesWithUserIdRequest) (*pb.FindAllEnabledServerNamesWithUserIdResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	servers, err := models.SharedServerDAO.FindAllEnabledServersWithUserId(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	serverNames := []string{}
	for _, server := range servers {
		if len(server.ServerNames) > 0 && server.ServerNames != "null" {
			serverNameConfigs := []*serverconfigs.ServerNameConfig{}
			err = json.Unmarshal([]byte(server.ServerNames), &serverNameConfigs)
			if err != nil {
				return nil, err
			}
			for _, config := range serverNameConfigs {
				if len(config.SubNames) == 0 {
					serverNames = append(serverNames, config.Name)
				} else {
					serverNames = append(serverNames, config.SubNames...)
				}
			}
		}
	}
	return &pb.FindAllEnabledServerNamesWithUserIdResponse{ServerNames: serverNames}, nil
}

// 查找服务基本信息
func (this *ServerService) FindEnabledUserServerBasic(ctx context.Context, req *pb.FindEnabledUserServerBasicRequest) (*pb.FindEnabledUserServerBasicResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	server, err := models.SharedServerDAO.FindEnabledServerBasic(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return &pb.FindEnabledUserServerBasicResponse{Server: nil}, nil
	}

	return &pb.FindEnabledUserServerBasicResponse{Server: &pb.Server{
		Id:          int64(server.Id),
		Name:        server.Name,
		Description: server.Description,
		IsOn:        server.IsOn == 1,
		Type:        server.Type,
	}}, nil
}

// 修改用户服务基本信息
func (this *ServerService) UpdateEnabledUserServerBasic(ctx context.Context, req *pb.UpdateEnabledUserServerBasicRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedServerDAO.UpdateUserServerBasic(tx, req.ServerId, req.Name)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
