package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/messageconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

type ServerService struct {
	BaseService
}

// CreateServer 创建服务
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

		// 集群
		nodeClusterId, err := models.SharedUserDAO.FindUserClusterId(tx, userId)
		if err != nil {
			return nil, err
		}
		if nodeClusterId > 0 {
			req.NodeClusterId = nodeClusterId
		}

		// 服务分组
		config, err := models.SharedSysSettingDAO.ReadUserServerConfig(tx)
		if err == nil && config.GroupId > 0 && !lists.ContainsInt64(req.ServerGroupIds, config.GroupId) {
			req.ServerGroupIds = append(req.ServerGroupIds, config.GroupId)
		}
	} else if req.UserId > 0 {
		// 集群
		nodeClusterId, err := models.SharedUserDAO.FindUserClusterId(tx, req.UserId)
		if err != nil {
			return nil, err
		}
		if nodeClusterId > 0 {
			req.NodeClusterId = nodeClusterId
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

	// 检查用户套餐
	if req.UserPlanId > 0 {
		userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, req.UserPlanId, nil)
		if err != nil {
			return nil, err
		}
		if userPlan == nil {
			return nil, errors.New("can not find user plan with id '" + types.String(req.UserPlanId) + "'")
		}
		if userId > 0 && int64(userPlan.UserId) != userId {
			return nil, errors.New("invalid user plan")
		}
		if req.UserId > 0 && int64(userPlan.UserId) != req.UserId {
			return nil, errors.New("invalid user plan")
		}

		// 套餐
		plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
		if err != nil {
			return nil, err
		}
		if plan == nil {
			return nil, errors.New("invalid plan: " + types.String(userPlan.PlanId))
		}
		if plan.ClusterId > 0 {
			req.NodeClusterId = int64(plan.ClusterId)
		}

		// 检查是否已经被别的服务所占用
		planServerId, err := models.SharedServerDAO.FindEnabledServerIdWithUserPlanId(tx, req.UserPlanId)
		if err != nil {
			return nil, err
		}
		if planServerId > 0 {
			return nil, errors.New("the user plan is used by another server '" + types.String(planServerId) + "'")
		}
	}

	serverId, err := models.SharedServerDAO.CreateServer(tx, req.AdminId, req.UserId, req.Type, req.Name, req.Description, serverNamesJSON, isAuditing, auditingServerNamesJSON, string(req.HttpJSON), string(req.HttpsJSON), string(req.TcpJSON), string(req.TlsJSON), string(req.UnixJSON), string(req.UdpJSON), req.WebId, req.ReverseProxyJSON, req.NodeClusterId, string(req.IncludeNodesJSON), string(req.ExcludeNodesJSON), req.ServerGroupIds, req.UserPlanId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateServerResponse{ServerId: serverId}, nil
}

// UpdateServerBasic 修改服务基本信息
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

	err = models.SharedServerDAO.UpdateServerBasic(tx, req.ServerId, req.Name, req.Description, req.NodeClusterId, req.IsOn, req.ServerGroupIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerIsOn 修改服务是否启用
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

// UpdateServerHTTP 修改HTTP服务
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

// UpdateServerHTTPS 修改HTTPS服务
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

// UpdateServerTCP 修改TCP服务
func (this *ServerService) UpdateServerTCP(ctx context.Context, req *pb.UpdateServerTCPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerTCP(tx, req.ServerId, req.TcpJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerTLS 修改TLS服务
func (this *ServerService) UpdateServerTLS(ctx context.Context, req *pb.UpdateServerTLSRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerDAO.UpdateServerTLS(tx, req.ServerId, req.TlsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerUnix 修改Unix服务
func (this *ServerService) UpdateServerUnix(ctx context.Context, req *pb.UpdateServerUnixRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
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

// UpdateServerUDP 修改UDP服务
func (this *ServerService) UpdateServerUDP(ctx context.Context, req *pb.UpdateServerUDPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
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

// UpdateServerWeb 修改Web服务
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

// UpdateServerReverseProxy 修改反向代理服务
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

// FindServerNames 查找服务的域名设置
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

	serverNamesJSON, isAuditing, auditingAt, auditingServerNamesJSON, auditingResultJSON, err := models.SharedServerDAO.FindServerServerNames(tx, req.ServerId)
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
		AuditingAt:              auditingAt,
		AuditingServerNamesJSON: auditingServerNamesJSON,
		AuditingResult:          auditingResult,
	}, nil
}

// UpdateServerNames 修改域名服务
func (this *ServerService) UpdateServerNames(ctx context.Context, req *pb.UpdateServerNamesRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查用户
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}

		// 是否需要审核
		globalConfig, err := models.SharedSysSettingDAO.ReadGlobalConfig(tx)
		if err != nil {
			return nil, err
		}
		if globalConfig != nil && globalConfig.HTTPAll.DomainAuditingIsOn {
			err = models.SharedServerDAO.UpdateAuditingServerNames(tx, req.ServerId, true, req.ServerNamesJSON)
			if err != nil {
				return nil, err
			}

			// 发送审核通知
			err = models.SharedMessageDAO.CreateMessage(tx, 0, 0, models.MessageTypeServerNamesRequireAuditing, models.MessageLevelWarning, "有新的网站域名需要审核", "有新的网站域名需要审核", maps.Map{
				"serverId": req.ServerId,
			}.AsJSON())

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

// UpdateServerNamesAuditing 审核服务的域名设置
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
			subject := "服务域名审核通过"
			msg := "服务域名审核通过"
			err = models.SharedMessageDAO.CreateMessage(tx, 0, userId, models.MessageTypeServerNamesAuditingSuccess, models.MessageLevelSuccess, subject, msg, maps.Map{
				"serverId": req.ServerId,
			}.AsJSON())
			if err != nil {
				return nil, err
			}
		} else {
			subject := "服务域名审核失败"
			msg := "服务域名审核失败，原因：" + req.AuditingResult.Reason
			err = models.SharedMessageDAO.CreateMessage(tx, 0, userId, models.MessageTypeServerNamesAuditingFailed, models.LevelError, subject, msg, maps.Map{
				"serverId": req.ServerId,
			}.AsJSON())
			if err != nil {
				return nil, err
			}
		}
	}

	return this.Success()
}

// UpdateServerDNS 修改服务的DNS相关设置
func (this *ServerService) UpdateServerDNS(ctx context.Context, req *pb.UpdateServerDNSRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedServerDAO.UpdateServerDNS(tx, req.ServerId, req.SupportCNAME)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// RegenerateServerCNAME 重新生成CNAME
func (this *ServerService) RegenerateServerCNAME(ctx context.Context, req *pb.RegenerateServerCNAMERequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	_, err = models.SharedServerDAO.GenerateServerDNSName(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledServersMatch 计算服务数量
func (this *ServerService) CountAllEnabledServersMatch(ctx context.Context, req *pb.CountAllEnabledServersMatchRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersMatch(tx, req.ServerGroupId, req.Keyword, req.UserId, req.NodeClusterId, types.Int8(req.AuditingFlag), utils.SplitStrings(req.ProtocolFamily, ","))
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledServersMatch 列出单页服务
func (this *ServerService) ListEnabledServersMatch(ctx context.Context, req *pb.ListEnabledServersMatchRequest) (*pb.ListEnabledServersMatchResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	servers, err := models.SharedServerDAO.ListEnabledServersMatch(tx, req.Offset, req.Size, req.ServerGroupId, req.Keyword, req.UserId, req.NodeClusterId, req.AuditingFlag, utils.SplitStrings(req.ProtocolFamily, ","))
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

		// 配置
		config, err := models.SharedServerDAO.ComposeServerConfig(tx, server, nil, false)
		if err != nil {
			return nil, err
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.Server{
			Id:                      int64(server.Id),
			IsOn:                    server.IsOn == 1,
			Type:                    server.Type,
			Config:                  configJSON,
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
			AuditingAt:              int64(server.AuditingAt),
			AuditingServerNamesJSON: []byte(server.AuditingServerNames),
			AuditingResult:          auditingResult,
			CreatedAt:               int64(server.CreatedAt),
			DnsName:                 server.DnsName,
			NodeCluster: &pb.NodeCluster{
				Id:   int64(server.ClusterId),
				Name: clusterName,
			},
			ServerGroups: pbGroups,
			User:         pbUser,
		})
	}

	return &pb.ListEnabledServersMatchResponse{Servers: result}, nil
}

// DeleteServer 禁用某服务
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

// FindEnabledServer 查找单个服务
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

	// 配置
	config, err := models.SharedServerDAO.ComposeServerConfig(tx, server, nil, userId > 0)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledServerResponse{Server: &pb.Server{
		Id:           int64(server.Id),
		IsOn:         server.IsOn == 1,
		Type:         server.Type,
		Name:         server.Name,
		Description:  server.Description,
		DnsName:      server.DnsName,
		SupportCNAME: server.SupportCNAME == 1,
		UserPlanId:   int64(server.UserPlanId),

		Config:           configJSON,
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
		ServerGroups: pbGroups,
		User:         pbUser,
	}}, nil
}

// FindEnabledServerConfig 查找服务配置
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

	config, err := models.SharedServerDAO.ComposeServerConfigWithServerId(tx, req.ServerId, false)
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

// FindEnabledServerType 查找服务的服务类型
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

// FindAndInitServerReverseProxyConfig 查找反向代理设置
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

	reverseProxyConfig, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, nil)
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

// FindAndInitServerWebConfig 初始化Web设置
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

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindAndInitServerWebConfigResponse{WebJSON: configJSON}, nil
}

// CountAllEnabledServersWithSSLCertId 计算使用某个SSL证书的服务数量
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

// FindAllEnabledServersWithSSLCertId 查找使用某个SSL证书的所有服务
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

// CountAllEnabledServersWithNodeClusterId 计算运行在某个集群上的所有服务数量
func (this *ServerService) CountAllEnabledServersWithNodeClusterId(ctx context.Context, req *pb.CountAllEnabledServersWithNodeClusterIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
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

// CountAllEnabledServersWithServerGroupId 计算使用某个分组的服务数量
func (this *ServerService) CountAllEnabledServersWithServerGroupId(ctx context.Context, req *pb.CountAllEnabledServersWithServerGroupIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersWithGroupId(tx, req.ServerGroupId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// NotifyServersChange 通知更新
func (this *ServerService) NotifyServersChange(ctx context.Context, _ *pb.NotifyServersChangeRequest) (*pb.NotifyServersChangeResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
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

// FindAllEnabledServersDNSWithNodeClusterId 取得某个集群下的所有服务相关的DNS
func (this *ServerService) FindAllEnabledServersDNSWithNodeClusterId(ctx context.Context, req *pb.FindAllEnabledServersDNSWithNodeClusterIdRequest) (*pb.FindAllEnabledServersDNSWithNodeClusterIdResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
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

	return &pb.FindAllEnabledServersDNSWithNodeClusterIdResponse{Servers: result}, nil
}

// FindEnabledServerDNS 查找单个服务的DNS信息
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

	supportCNAME, err := models.SharedServerDAO.FindServerSupportCNAME(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	var pbDomain *pb.DNSDomain = nil
	if clusterId > 0 {
		clusterDNS, err := models.SharedNodeClusterDAO.FindClusterDNSInfo(tx, clusterId, nil)
		if err != nil {
			return nil, err
		}
		if clusterDNS != nil {
			domainId := int64(clusterDNS.DnsDomainId)
			if domainId > 0 {
				domain, err := dns.SharedDNSDomainDAO.FindEnabledDNSDomain(tx, domainId, nil)
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
		DnsName:      dnsName,
		Domain:       pbDomain,
		SupportCNAME: supportCNAME,
	}, nil
}

// CheckUserServer 检查服务是否属于某个用户
func (this *ServerService) CheckUserServer(ctx context.Context, req *pb.CheckUserServerRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx)
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

// FindAllEnabledServerNamesWithUserId 查找一个用户下的所有域名列表
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

// FindEnabledUserServerBasic 查找服务基本信息
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

	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(server.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledUserServerBasicResponse{Server: &pb.Server{
		Id:          int64(server.Id),
		Name:        server.Name,
		Description: server.Description,
		IsOn:        server.IsOn == 1,
		Type:        server.Type,
		NodeCluster: &pb.NodeCluster{
			Id:   int64(server.ClusterId),
			Name: clusterName,
		},
	}}, nil
}

// UpdateEnabledUserServerBasic 修改用户服务基本信息
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

// UploadServerHTTPRequestStat 上传待统计数据
func (this *ServerService) UploadServerHTTPRequestStat(ctx context.Context, req *pb.UploadServerHTTPRequestStatRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	month := req.Month
	if len(month) == 0 {
		month = timeutil.Format("Ym")
	}

	day := req.Day
	if len(day) == 0 {
		day = timeutil.Format("Ymd")
	}

	// 区域
	for _, result := range req.RegionCities {
		// IP => 地理位置
		err := func() error {
			// 区域
			if len(result.CountryName) > 0 {
				countryId, err := regions.SharedRegionCountryDAO.FindCountryIdWithNameCacheable(tx, result.CountryName)
				if err != nil {
					return err
				}
				if countryId > 0 {
					countryKey := fmt.Sprintf("%d@%d@%s", result.ServerId, countryId, day)
					serverStatLocker.Lock()
					stat, ok := serverHTTPCountryStatMap[countryKey]
					if !ok {
						stat = &TrafficStat{}
						serverHTTPCountryStatMap[countryKey] = stat
					}
					stat.CountRequests += result.CountRequests
					stat.Bytes += result.Bytes
					stat.CountAttackRequests += result.CountAttackRequests
					stat.AttackBytes += result.AttackBytes
					serverStatLocker.Unlock()

					// 省份
					if len(result.ProvinceName) > 0 {
						provinceId, err := regions.SharedRegionProvinceDAO.FindProvinceIdWithNameCacheable(tx, countryId, result.ProvinceName)
						if err != nil {
							return err
						}
						if provinceId > 0 {
							key := fmt.Sprintf("%d@%d@%s", result.ServerId, provinceId, month)
							serverStatLocker.Lock()
							serverHTTPProvinceStatMap[key] += result.CountRequests
							serverStatLocker.Unlock()

							// 城市
							if len(result.CityName) > 0 {
								cityId, err := regions.SharedRegionCityDAO.FindCityIdWithNameCacheable(tx, provinceId, result.CityName)
								if err != nil {
									return err
								}
								if cityId > 0 {
									key := fmt.Sprintf("%d@%d@%s", result.ServerId, cityId, month)
									serverStatLocker.Lock()
									serverHTTPCityStatMap[key] += result.CountRequests
									serverStatLocker.Unlock()
								}
							}

						}
					}
				}
			}

			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// 运营商
	for _, result := range req.RegionProviders {
		// IP => 地理位置
		err := func() error {
			if len(result.Name) == 0 {
				return nil
			}
			providerId, err := regions.SharedRegionProviderDAO.FindProviderIdWithNameCacheable(tx, result.Name)
			if err != nil {
				return err
			}
			if providerId > 0 {
				key := fmt.Sprintf("%d@%d@%s", result.ServerId, providerId, month)
				serverStatLocker.Lock()
				serverHTTPProviderStatMap[key] += result.Count
				serverStatLocker.Unlock()
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// OS
	for _, result := range req.Systems {
		err := func() error {
			if len(result.Name) == 0 {
				return nil
			}

			systemId, err := models.SharedClientSystemDAO.FindSystemIdWithNameCacheable(tx, result.Name)
			if err != nil {
				return err
			}
			if systemId == 0 {
				systemId, err = models.SharedClientSystemDAO.CreateSystem(tx, result.Name)
				if err != nil {
					return err
				}
			}
			key := fmt.Sprintf("%d@%d@%s@%s", result.ServerId, systemId, result.Version, month)
			serverStatLocker.Lock()
			serverHTTPSystemStatMap[key] += result.Count
			serverStatLocker.Unlock()
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// Browser
	for _, result := range req.Browsers {
		err := func() error {
			if len(result.Name) == 0 {
				return nil
			}

			browserId, err := models.SharedClientBrowserDAO.FindBrowserIdWithNameCacheable(tx, result.Name)
			if err != nil {
				return err
			}
			if browserId == 0 {
				browserId, err = models.SharedClientBrowserDAO.CreateBrowser(tx, result.Name)
				if err != nil {
					return err
				}
			}
			key := fmt.Sprintf("%d@%d@%s@%s", result.ServerId, browserId, result.Version, month)
			serverStatLocker.Lock()
			serverHTTPBrowserStatMap[key] += result.Count
			serverStatLocker.Unlock()
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// 防火墙
	for _, result := range req.HttpFirewallRuleGroups {
		err := func() error {
			if result.HttpFirewallRuleGroupId <= 0 {
				return nil
			}
			key := fmt.Sprintf("%d@%d@%s@%s", result.ServerId, result.HttpFirewallRuleGroupId, result.Action, day)
			serverStatLocker.Lock()
			serverHTTPFirewallRuleGroupStatMap[key] += result.Count
			serverStatLocker.Unlock()
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// CheckServerNameDuplicationInNodeCluster 检查域名是否已经存在
func (this *ServerService) CheckServerNameDuplicationInNodeCluster(ctx context.Context, req *pb.CheckServerNameDuplicationInNodeClusterRequest) (*pb.CheckServerNameDuplicationInNodeClusterResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(req.ServerNames) == 0 {
		return &pb.CheckServerNameDuplicationInNodeClusterResponse{DuplicatedServerNames: nil}, nil
	}

	var tx = this.NullTx()

	duplicatedServerNames := []string{}
	for _, serverName := range req.ServerNames {
		exist, err := models.SharedServerDAO.ExistServerNameInCluster(tx, req.NodeClusterId, serverName, req.ExcludeServerId)
		if err != nil {
			return nil, err
		}
		if exist {
			duplicatedServerNames = append(duplicatedServerNames, serverName)
		}
	}

	return &pb.CheckServerNameDuplicationInNodeClusterResponse{DuplicatedServerNames: duplicatedServerNames}, nil
}

// FindLatestServers 查找最近访问的服务
func (this *ServerService) FindLatestServers(ctx context.Context, req *pb.FindLatestServersRequest) (*pb.FindLatestServersResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	servers, err := models.SharedServerDAO.FindLatestServers(tx, req.Size)
	if err != nil {
		return nil, err
	}
	pbServers := []*pb.Server{}
	for _, server := range servers {
		pbServers = append(pbServers, &pb.Server{
			Id:   int64(server.Id),
			Name: server.Name,
		})
	}
	return &pb.FindLatestServersResponse{Servers: pbServers}, nil
}

// FindNearbyServers 查找某个服务附近的服务
func (this *ServerService) FindNearbyServers(ctx context.Context, req *pb.FindNearbyServersRequest) (*pb.FindNearbyServersResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 查询服务的Group
	groupIds, err := models.SharedServerDAO.FindServerGroupIds(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if len(groupIds) > 0 {
		var pbGroups = []*pb.FindNearbyServersResponse_GroupInfo{}
		for _, groupId := range groupIds {
			group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(tx, groupId)
			if err != nil {
				return nil, err
			}
			if group == nil {
				continue
			}

			var pbGroup = &pb.FindNearbyServersResponse_GroupInfo{
				Name: group.Name,
			}
			servers, err := models.SharedServerDAO.FindNearbyServersInGroup(tx, groupId, req.ServerId, 10)
			if err != nil {
				return nil, err
			}
			for _, server := range servers {
				pbGroup.Servers = append(pbGroup.Servers, &pb.Server{
					Id:   int64(server.Id),
					Name: server.Name,
					IsOn: server.IsOn == 1,
				})
			}
			pbGroups = append(pbGroups, pbGroup)
		}

		if len(pbGroups) > 0 {
			return &pb.FindNearbyServersResponse{
				Scope:  "group",
				Groups: pbGroups,
			}, nil
		}
	}

	// 集群
	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	servers, err := models.SharedServerDAO.FindNearbyServersInCluster(tx, clusterId, req.ServerId, 10)
	if err != nil {
		return nil, err
	}
	if len(servers) == 0 {
		return &pb.FindNearbyServersResponse{
			Scope:  "cluster",
			Groups: nil,
		}, nil
	}

	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, clusterId)
	if err != nil {
		return nil, err
	}
	var pbGroup = &pb.FindNearbyServersResponse_GroupInfo{
		Name: clusterName,
	}
	for _, server := range servers {
		pbGroup.Servers = append(pbGroup.Servers, &pb.Server{
			Id:   int64(server.Id),
			Name: server.Name,
			IsOn: server.IsOn == 1,
		})
	}

	return &pb.FindNearbyServersResponse{
		Scope:  "cluster",
		Groups: []*pb.FindNearbyServersResponse_GroupInfo{pbGroup},
	}, nil
}

// PurgeServerCache 清除缓存
func (this *ServerService) PurgeServerCache(ctx context.Context, req *pb.PurgeServerCacheRequest) (*pb.PurgeServerCacheResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		// 检查是否为节点
		_, err = this.ValidateNode(ctx)
		if err != nil {
			return nil, err
		}
	}

	if len(req.Domains) == 0 {
		return nil, errors.New("'domains' field is required")
	}

	if len(req.Keys) == 0 && len(req.Prefixes) == 0 {
		return &pb.PurgeServerCacheResponse{IsOk: true}, nil
	}

	var tx = this.NullTx()
	var cacheMap = utils.NewCacheMap()
	var purgeResponse = &pb.PurgeServerCacheResponse{}

	for _, domain := range req.Domains {
		servers, err := models.SharedServerDAO.FindAllEnabledServersWithDomain(tx, domain)
		if err != nil {
			return nil, err
		}

		for _, server := range servers {
			clusterId := int64(server.ClusterId)
			if clusterId > 0 {
				nodeIds, err := models.SharedNodeDAO.FindAllEnabledNodeIdsWithClusterId(tx, clusterId)
				if err != nil {
					return nil, err
				}

				cachePolicyId, err := models.SharedNodeClusterDAO.FindClusterHTTPCachePolicyId(tx, clusterId, cacheMap)
				if err != nil {
					return nil, err
				}
				if cachePolicyId == 0 {
					continue
				}

				cachePolicy, err := models.SharedHTTPCachePolicyDAO.ComposeCachePolicy(tx, cachePolicyId, cacheMap)
				if err != nil {
					return nil, err
				}
				if cachePolicy == nil {
					continue
				}
				cachePolicyJSON, err := json.Marshal(cachePolicy)
				if err != nil {
					return nil, err
				}

				for _, nodeId := range nodeIds {
					msg := &messageconfigs.PurgeCacheMessage{
						CachePolicyJSON: cachePolicyJSON,
					}
					if len(req.Prefixes) > 0 {
						msg.Type = messageconfigs.PurgeCacheMessageTypeDir
						msg.Keys = req.Prefixes
					} else {
						msg.Type = messageconfigs.PurgeCacheMessageTypeFile
						msg.Keys = req.Keys
					}
					msgJSON, err := json.Marshal(msg)
					if err != nil {
						return nil, err
					}

					resp, err := SendCommandToNode(nodeId, NextCommandRequestId(), messageconfigs.MessageCodePurgeCache, msgJSON, 10, false)
					if err != nil {
						return nil, err
					}
					if !resp.IsOk {
						purgeResponse.IsOk = false
						purgeResponse.Message = resp.Message
						return purgeResponse, nil
					}
				}
			}
		}
	}

	purgeResponse.IsOk = true

	return purgeResponse, nil
}

// FindEnabledServerTrafficLimit 查找流量限制
func (this *ServerService) FindEnabledServerTrafficLimit(ctx context.Context, req *pb.FindEnabledServerTrafficLimitRequest) (*pb.FindEnabledServerTrafficLimitResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查用户权限

	var tx = this.NullTx()
	limitConfig, err := models.SharedServerDAO.FindServerTrafficLimitConfig(tx, req.ServerId, nil)
	if err != nil {
		return nil, err
	}
	limitConfigJSON, err := json.Marshal(limitConfig)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledServerTrafficLimitResponse{
		TrafficLimitJSON: limitConfigJSON,
	}, nil
}

// UpdateServerTrafficLimit 设置流量限制
func (this *ServerService) UpdateServerTrafficLimit(ctx context.Context, req *pb.UpdateServerTrafficLimitRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var config = &serverconfigs.TrafficLimitConfig{}
	err = json.Unmarshal(req.TrafficLimitJSON, config)
	if err != nil {
		return nil, err
	}

	err = models.SharedServerDAO.UpdateServerTrafficLimitConfig(tx, req.ServerId, config)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateServerUserPlan 修改服务套餐
func (this *ServerService) UpdateServerUserPlan(ctx context.Context, req *pb.UpdateServerUserPlanRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		// 检查服务
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 检查套餐
	if req.UserPlanId < 0 {
		req.UserPlanId = 0
	}

	// 检查是否有变化
	oldUserPlanId, err := models.SharedServerDAO.FindServerUserPlanId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if req.UserPlanId == oldUserPlanId {
		return this.Success()
	}

	if req.UserPlanId > 0 {
		userId, err := models.SharedServerDAO.FindServerUserId(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
		if userId == 0 {
			return nil, errors.New("the server is not belong to any user")
		}

		userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, req.UserPlanId, nil)
		if err != nil {
			return nil, err
		}
		if userPlan == nil {
			return nil, errors.New("can not find user plan with id '" + types.String(req.UserPlanId) + "'")
		}
		if int64(userPlan.UserId) != userId {
			return nil, errors.New("can not find user plan with id '" + types.String(req.UserPlanId) + "'")
		}

		// 检查是否已经被别的服务所使用
		serverId, err := models.SharedServerDAO.FindEnabledServerIdWithUserPlanId(tx, req.UserPlanId)
		if err != nil {
			return nil, err
		}
		if serverId > 0 && serverId != req.ServerId {
			return nil, errors.New("the user plan is used by other server")
		}
	}

	err = models.SharedServerDAO.UpdateServerUserPlanId(tx, req.ServerId, req.UserPlanId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindServerUserPlan 获取服务套餐信息
func (this *ServerService) FindServerUserPlan(ctx context.Context, req *pb.FindServerUserPlanRequest) (*pb.FindServerUserPlanResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		// 检查服务
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	userPlanId, err := models.SharedServerDAO.FindServerUserPlanId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if userPlanId <= 0 {
		return &pb.FindServerUserPlanResponse{UserPlan: nil}, nil
	}

	userPlan, err := models.SharedUserPlanDAO.FindEnabledUserPlan(tx, userPlanId, nil)
	if err != nil {
		return nil, err
	}
	if userPlan == nil {
		return &pb.FindServerUserPlanResponse{UserPlan: nil}, nil
	}

	plan, err := models.SharedPlanDAO.FindEnabledPlan(tx, int64(userPlan.PlanId))
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return &pb.FindServerUserPlanResponse{UserPlan: nil}, nil
	}

	return &pb.FindServerUserPlanResponse{
		UserPlan: &pb.UserPlan{
			Id:     int64(userPlan.Id),
			UserId: int64(userPlan.UserId),
			PlanId: int64(userPlan.PlanId),
			Name:   userPlan.Name,
			IsOn:   userPlan.IsOn == 1,
			DayTo:  userPlan.DayTo,
			User:   nil,
			Plan: &pb.Plan{
				Id:        int64(plan.Id),
				Name:      plan.Name,
				PriceType: plan.PriceType,
			},
		},
	}, nil
}
