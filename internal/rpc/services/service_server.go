package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/clients"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"strings"
)

type ServerService struct {
	BaseService
}

// CreateServer 创建服务
func (this *ServerService) CreateServer(ctx context.Context, req *pb.CreateServerRequest) (*pb.CreateServerResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

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
				err := models.SharedSSLPolicyDAO.CheckUserPolicy(tx, userId, httpsConfig.SSLPolicyRef.SSLPolicyId)
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
				err := models.SharedSSLPolicyDAO.CheckUserPolicy(tx, userId, tlsConfig.SSLPolicyRef.SSLPolicyId)
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
		for _, groupId := range req.ServerGroupIds {
			err := models.SharedServerGroupDAO.CheckUserGroup(tx, userId, groupId)
			if err != nil {
				return nil, err
			}
		}

		// 增加默认分组
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

	serverId, err := models.SharedServerDAO.CreateServer(tx, req.AdminId, req.UserId, req.Type, req.Name, req.Description, serverNamesJSON, isAuditing, auditingServerNamesJSON, req.HttpJSON, req.HttpsJSON, req.TcpJSON, req.TlsJSON, req.UnixJSON, req.UdpJSON, req.WebId, req.ReverseProxyJSON, req.NodeClusterId, req.IncludeNodesJSON, req.ExcludeNodesJSON, req.ServerGroupIds, req.UserPlanId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateServerResponse{ServerId: serverId}, nil
}

// UpdateServerBasic 修改服务基本信息
func (this *ServerService) UpdateServerBasic(ctx context.Context, req *pb.UpdateServerBasicRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	var tx = this.NullTx()

	// 查询老的节点信息
	server, err := models.SharedServerDAO.FindEnabledServer(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("can not find server")
	}

	err = models.SharedServerDAO.UpdateServerBasic(tx, req.ServerId, req.Name, req.Description, req.NodeClusterId, req.KeepOldConfigs, req.IsOn, req.ServerGroupIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerGroupIds 修改服务所在分组
func (this *ServerService) UpdateServerGroupIds(ctx context.Context, req *pb.UpdateServerGroupIdsRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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

	// 检查分组IDs
	var serverGroupIds = []int64{}
	for _, groupId := range req.ServerGroupIds {
		if userId > 0 {
			err = models.SharedServerGroupDAO.CheckUserGroup(tx, userId, groupId)
			if err != nil {
				return nil, err
			}
		} else {
			b, err := models.SharedServerGroupDAO.ExistsGroup(tx, groupId)
			if err != nil {
				return nil, err
			}
			if !b {
				continue
			}
		}
		serverGroupIds = append(serverGroupIds, groupId)
	}

	// 增加默认分组
	if userId > 0 {
		config, err := models.SharedSysSettingDAO.ReadUserServerConfig(tx)
		if err == nil && config.GroupId > 0 && !lists.ContainsInt64(serverGroupIds, config.GroupId) {
			serverGroupIds = append(serverGroupIds, config.GroupId)
		}
	}

	// 修改
	err = models.SharedServerDAO.UpdateServerGroupIds(tx, req.ServerId, serverGroupIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerIsOn 修改服务是否启用
func (this *ServerService) UpdateServerIsOn(ctx context.Context, req *pb.UpdateServerIsOnRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
	err = models.SharedServerDAO.UpdateServerIsOn(tx, req.ServerId, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateServerHTTP 修改HTTP服务
func (this *ServerService) UpdateServerHTTP(ctx context.Context, req *pb.UpdateServerHTTPRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()

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
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	var tx = this.NullTx()

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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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

	var tx = this.NullTx()

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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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

	// 修改配置
	err = models.SharedServerDAO.UpdateServerReverseProxy(tx, req.ServerId, req.ReverseProxyJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindServerNames 查找服务的域名设置
func (this *ServerService) FindServerNames(ctx context.Context, req *pb.FindServerNamesRequest) (*pb.FindServerNamesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 转换为小写
	var serverNameConfigs = []*serverconfigs.ServerNameConfig{}
	if len(req.ServerNamesJSON) > 0 {
		err = json.Unmarshal(req.ServerNamesJSON, &serverNameConfigs)
		if err != nil {
			return nil, err
		}
		if len(serverNameConfigs) > 0 {
			for _, serverName := range serverNameConfigs {
				serverName.Normalize()
			}
			req.ServerNamesJSON, err = json.Marshal(serverNameConfigs)
			if err != nil {
				return nil, err
			}
		}
	}

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
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.AuditingResult == nil {
		return nil, errors.New("'result' should not be nil")
	}

	var tx = this.NullTx()

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
	_, err := this.ValidateAdmin(ctx)
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

// RegenerateServerDNSName 重新生成CNAME
func (this *ServerService) RegenerateServerDNSName(ctx context.Context, req *pb.RegenerateServerDNSNameRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
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

// UpdateServerDNSName 修改服务的CNAME
func (this *ServerService) UpdateServerDNSName(ctx context.Context, req *pb.UpdateServerDNSNameRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var dnsName = req.DnsName

	if req.ServerId <= 0 {
		return nil, errors.New("invalid 'serverId'")
	}

	if len(dnsName) == 0 {
		return nil, errors.New("'dnsName' must not be empty")
	}

	// 处理格式
	dnsName = strings.ToLower(dnsName)
	const maxLen = 30
	if len(dnsName) > maxLen {
		return nil, errors.New("'dnsName' too long than " + types.String(maxLen))
	}
	if !regexp.MustCompile(`^[a-z0-9]{1,` + types.String(maxLen) + `}$`).MatchString(dnsName) {
		return nil, errors.New("invalid 'dnsName': contains invalid character(s)")
	}

	// 检查是否被使用
	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, errors.New("the server is not belong to any cluster")
	}

	serverId, err := models.SharedServerDAO.FindServerIdWithDNSName(tx, clusterId, dnsName)
	if err != nil {
		return nil, err
	}
	if serverId > 0 && serverId != req.ServerId {
		return nil, errors.New("the 'dnsName': " + dnsName + " has already been used")
	}

	err = models.SharedServerDAO.UpdateServerDNSName(tx, req.ServerId, dnsName)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindServerIdWithDNSName 使用CNAME查找服务
func (this *ServerService) FindServerIdWithDNSName(ctx context.Context, req *pb.FindServerIdWithDNSNameRequest) (*pb.FindServerIdWithDNSNameResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.DnsName) == 0 {
		return nil, errors.New("'dnsName' must not be empty")
	}

	var tx = this.NullTx()
	serverId, err := models.SharedServerDAO.FindServerIdWithDNSName(tx, req.NodeClusterId, req.DnsName)
	if err != nil {
		return nil, err
	}

	return &pb.FindServerIdWithDNSNameResponse{
		ServerId: serverId,
	}, nil
}

// CountAllEnabledServersMatch 计算服务数量
func (this *ServerService) CountAllEnabledServersMatch(ctx context.Context, req *pb.CountAllEnabledServersMatchRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersMatch(tx, req.ServerGroupId, req.Keyword, req.UserId, req.NodeClusterId, types.Int8(req.AuditingFlag), utils.SplitStrings(req.ProtocolFamily, ","))
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledServersMatch 列出单页服务
func (this *ServerService) ListEnabledServersMatch(ctx context.Context, req *pb.ListEnabledServersMatchRequest) (*pb.ListEnabledServersMatchResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	var fromUser = false
	if userId > 0 {
		fromUser = true
		req.UserId = userId
	}

	var order = ""
	if req.TrafficOutAsc {
		order = "trafficOutAsc"
	} else if req.TrafficOutDesc {
		order = "trafficOutDesc"
	}

	servers, err := models.SharedServerDAO.ListEnabledServersMatch(tx, req.Offset, req.Size, req.ServerGroupId, req.Keyword, req.UserId, req.NodeClusterId, req.AuditingFlag, utils.SplitStrings(req.ProtocolFamily, ","), order)
	if err != nil {
		return nil, err
	}
	var result = []*pb.Server{}
	for _, server := range servers {
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(server.ClusterId))
		if err != nil {
			return nil, err
		}

		// 分组信息
		var pbGroups = []*pb.ServerGroup{}
		if models.IsNotNull(server.GroupIds) {
			var groupIds = []int64{}
			err = json.Unmarshal(server.GroupIds, &groupIds)
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
					Id:     int64(group.Id),
					Name:   group.Name,
					UserId: int64(group.UserId),
				})
			}
		}

		// 用户
		var pbUser *pb.User = nil
		if !fromUser {
			user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, int64(server.UserId))
			if err != nil {
				return nil, err
			}
			if user != nil {
				pbUser = &pb.User{
					Id:       int64(user.Id),
					Fullname: user.Fullname,
				}
			}
		}

		// 审核结果
		var auditingResult = &pb.ServerNameAuditingResult{}
		if len(server.AuditingResult) > 0 {
			err = json.Unmarshal(server.AuditingResult, auditingResult)
			if err != nil {
				return nil, err
			}
		} else {
			auditingResult.IsOk = true
		}

		// 配置
		config, err := models.SharedServerDAO.ComposeServerConfig(tx, server, req.IgnoreSSLCerts, nil, nil, false, true)
		if err != nil {
			return nil, err
		}
		var countServerNames int32 = 0
		for _, serverName := range config.ServerNames {
			if len(serverName.SubNames) > 0 {
				countServerNames += int32(len(serverName.SubNames))
			} else {
				countServerNames++
			}
		}
		if req.IgnoreServerNames && len(config.ServerNames) > 0 {
			config.ServerNames = config.ServerNames[:1]
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}

		// 忽略信息
		if req.IgnoreServerNames {
			server.ServerNames = nil
		}

		result = append(result, &pb.Server{
			Id:                      int64(server.Id),
			IsOn:                    server.IsOn,
			Type:                    server.Type,
			Config:                  configJSON,
			Name:                    server.Name,
			CountServerNames:        countServerNames,
			Description:             server.Description,
			HttpJSON:                server.Http,
			HttpsJSON:               server.Https,
			TcpJSON:                 server.Tcp,
			TlsJSON:                 server.Tls,
			UnixJSON:                server.Unix,
			UdpJSON:                 server.Udp,
			IncludeNodes:            server.IncludeNodes,
			ExcludeNodes:            server.ExcludeNodes,
			ServerNamesJSON:         server.ServerNames,
			IsAuditing:              server.IsAuditing,
			AuditingAt:              int64(server.AuditingAt),
			AuditingServerNamesJSON: server.AuditingServerNames,
			AuditingResult:          auditingResult,
			CreatedAt:               int64(server.CreatedAt),
			DnsName:                 server.DnsName,
			UserPlanId:              int64(server.UserPlanId),
			NodeCluster: &pb.NodeCluster{
				Id:   int64(server.ClusterId),
				Name: clusterName,
			},
			ServerGroups:   pbGroups,
			UserId:         int64(server.UserId),
			User:           pbUser,
			BandwidthTime:  server.BandwidthTime,
			BandwidthBytes: int64(server.BandwidthBytes),
		})
	}

	return &pb.ListEnabledServersMatchResponse{Servers: result}, nil
}

// DeleteServer 禁用某服务
func (this *ServerService) DeleteServer(ctx context.Context, req *pb.DeleteServerRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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
	var pbGroups = []*pb.ServerGroup{}
	if len(server.GroupIds) > 0 {
		var groupIds = []int64{}
		err = json.Unmarshal(server.GroupIds, &groupIds)
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
				Id:     int64(group.Id),
				Name:   group.Name,
				UserId: int64(group.UserId),
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
	config, err := models.SharedServerDAO.ComposeServerConfig(tx, server, req.IgnoreSSLCerts, nil, nil, userId > 0, false)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledServerResponse{Server: &pb.Server{
		Id:           int64(server.Id),
		IsOn:         server.IsOn,
		Type:         server.Type,
		Name:         server.Name,
		Description:  server.Description,
		DnsName:      server.DnsName,
		SupportCNAME: server.SupportCNAME == 1,
		UserPlanId:   int64(server.UserPlanId),

		Config:           configJSON,
		ServerNamesJSON:  server.ServerNames,
		HttpJSON:         server.Http,
		HttpsJSON:        server.Https,
		TcpJSON:          server.Tcp,
		TlsJSON:          server.Tls,
		UnixJSON:         server.Unix,
		UdpJSON:          server.Udp,
		WebId:            int64(server.WebId),
		ReverseProxyJSON: server.ReverseProxy,

		IncludeNodes: server.IncludeNodes,
		ExcludeNodes: server.ExcludeNodes,
		CreatedAt:    int64(server.CreatedAt),
		NodeCluster: &pb.NodeCluster{
			Id:   int64(server.ClusterId),
			Name: clusterName,
		},
		ServerGroups: pbGroups,
		UserId:       int64(server.UserId),
		User:         pbUser,
	}}, nil
}

// FindEnabledServerConfig 查找服务配置
func (this *ServerService) FindEnabledServerConfig(ctx context.Context, req *pb.FindEnabledServerConfigRequest) (*pb.FindEnabledServerConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedServerDAO.ComposeServerConfigWithServerId(tx, req.ServerId, false, false)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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

	reverseProxyConfig, err := models.SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, reverseProxyRef.ReverseProxyId, nil, nil)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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

	config, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, false, false, nil, nil)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	if userId > 0 {
		// TODO 校验权限
	}

	var tx = this.NullTx()

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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}

	var tx = this.NullTx()

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
			IsOn: server.IsOn,
			Type: server.Type,
		})
	}
	return &pb.FindAllEnabledServersWithSSLCertIdResponse{Servers: result}, nil
}

// CountAllEnabledServersWithNodeClusterId 计算运行在某个集群上的所有服务数量
func (this *ServerService) CountAllEnabledServersWithNodeClusterId(ctx context.Context, req *pb.CountAllEnabledServersWithNodeClusterIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersWithNodeClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountAllEnabledServersWithServerGroupId 计算使用某个分组的服务数量
func (this *ServerService) CountAllEnabledServersWithServerGroupId(ctx context.Context, req *pb.CountAllEnabledServersWithServerGroupIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedServerDAO.CountAllEnabledServersWithGroupId(tx, req.ServerGroupId, userId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// NotifyServersChange 通知更新
func (this *ServerService) NotifyServersChange(ctx context.Context, _ *pb.NotifyServersChangeRequest) (*pb.NotifyServersChangeResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllEnabledServerNamesWithUserId 查找一个用户下的所有域名列表
func (this *ServerService) FindAllEnabledServerNamesWithUserId(ctx context.Context, req *pb.FindAllEnabledServerNamesWithUserIdRequest) (*pb.FindAllEnabledServerNamesWithUserIdResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	servers, err := models.SharedServerDAO.FindAllBasicServersWithUserId(tx, req.UserId)
	if err != nil {
		return nil, err
	}
	serverNames := []string{}
	for _, server := range servers {
		if models.IsNotNull(server.ServerNames) {
			serverNameConfigs := []*serverconfigs.ServerNameConfig{}
			err = json.Unmarshal(server.ServerNames, &serverNameConfigs)
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

// FindAllUserServers 查找一个用户下的所有服务
func (this *ServerService) FindAllUserServers(ctx context.Context, req *pb.FindAllUserServersRequest) (*pb.FindAllUserServersResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	servers, err := models.SharedServerDAO.FindAllBasicServersWithUserId(tx, req.UserId)
	if err != nil {
		return nil, err
	}

	var pbServers = []*pb.Server{}
	for _, server := range servers {
		pbServers = append(pbServers, &pb.Server{
			Id:              int64(server.Id),
			Name:            server.Name,
			FirstServerName: server.FirstServerName(),
			IsOn:            server.IsOn,
		})
	}

	return &pb.FindAllUserServersResponse{
		Servers: pbServers,
	}, nil
}

// ComposeAllUserServersConfig 查找某个用户下的服务配置
func (this *ServerService) ComposeAllUserServersConfig(ctx context.Context, req *pb.ComposeAllUserServersConfigRequest) (*pb.ComposeAllUserServersConfigResponse, error) {
	_, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	servers, err := models.SharedServerDAO.FindAllAvailableServersWithUserId(tx, req.UserId)
	if err != nil {
		return nil, err
	}

	var configs = []*serverconfigs.ServerConfig{}
	var cacheMap = utils.NewCacheMap()
	for _, server := range servers {
		config, err := models.SharedServerDAO.ComposeServerConfig(tx, server, false, nil, cacheMap, true, false)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	configsJSON, err := json.Marshal(configs)
	if err != nil {
		return nil, err
	}

	return &pb.ComposeAllUserServersConfigResponse{
		ServersConfigJSON: configsJSON,
	}, nil
}

// FindEnabledUserServerBasic 查找服务基本信息
func (this *ServerService) FindEnabledUserServerBasic(ctx context.Context, req *pb.FindEnabledUserServerBasicRequest) (*pb.FindEnabledUserServerBasicResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
		IsOn:        server.IsOn,
		Type:        server.Type,
		UserId:      int64(server.UserId),
		NodeCluster: &pb.NodeCluster{
			Id:   int64(server.ClusterId),
			Name: clusterName,
		},
	}}, nil
}

// UpdateEnabledUserServerBasic 修改用户服务基本信息
func (this *ServerService) UpdateEnabledUserServerBasic(ctx context.Context, req *pb.UpdateEnabledUserServerBasicRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
			if result.CountryId > 0 {
				var countryKey = fmt.Sprintf("%d@%d@%s", result.ServerId, result.CountryId, day)
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
				if result.ProvinceId > 0 {
					var provinceKey = fmt.Sprintf("%d@%d@%s", result.ServerId, result.ProvinceId, month)
					serverStatLocker.Lock()
					serverHTTPProvinceStatMap[provinceKey] += result.CountRequests
					serverStatLocker.Unlock()

					// 城市
					if result.CityId > 0 {
						var cityKey = fmt.Sprintf("%d@%d@%s", result.ServerId, result.CityId, month)
						serverStatLocker.Lock()
						serverHTTPCityStatMap[cityKey] += result.CountRequests
						serverStatLocker.Unlock()
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
			if result.ProviderId > 0 {
				var providerKey = fmt.Sprintf("%d@%d@%s", result.ServerId, result.ProviderId, month)
				serverStatLocker.Lock()
				serverHTTPProviderStatMap[providerKey] += result.Count
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

			systemId, err := models.SharedFormalClientSystemDAO.FindSystemIdWithNameCacheable(tx, result.Name)
			if err != nil {
				return err
			}
			if systemId == 0 {
				err = clients.SharedClientSystemDAO.CreateSystemIfNotExists(tx, result.Name)
				if err != nil {
					return err
				}

				// 直接返回不再进行操作
				return nil
			}
			var key = fmt.Sprintf("%d@%d@%s@%s", result.ServerId, systemId, result.Version, month)
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

			browserId, err := models.SharedFormalClientBrowserDAO.FindBrowserIdWithNameCacheable(tx, result.Name)
			if err != nil {
				return err
			}
			if browserId == 0 {
				err = clients.SharedClientBrowserDAO.CreateBrowserIfNotExists(tx, result.Name)
				if err != nil {
					return err
				}

				// 直接返回不再进行操作
				return nil
			}
			var key = fmt.Sprintf("%d@%d@%s@%s", result.ServerId, browserId, result.Version, month)
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
			var key = fmt.Sprintf("%d@%d@%s@%s", result.ServerId, result.HttpFirewallRuleGroupId, result.Action, day)
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
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if len(req.ServerNames) == 0 {
		return &pb.CheckServerNameDuplicationInNodeClusterResponse{DuplicatedServerNames: nil}, nil
	}

	var tx = this.NullTx()

	var duplicatedServerNames = []string{}
	for _, serverName := range req.ServerNames {
		exist, err := models.SharedServerDAO.ExistServerNameInCluster(tx, req.NodeClusterId, serverName, req.ExcludeServerId, req.SupportWildcard)
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
	_, err := this.ValidateAdmin(ctx)
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
	_, err := this.ValidateAdmin(ctx)
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
					IsOn: server.IsOn,
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
			IsOn: server.IsOn,
		})
	}

	return &pb.FindNearbyServersResponse{
		Scope:  "cluster",
		Groups: []*pb.FindNearbyServersResponse_GroupInfo{pbGroup},
	}, nil
}

// PurgeServerCache 清除缓存
func (this *ServerService) PurgeServerCache(ctx context.Context, req *pb.PurgeServerCacheRequest) (*pb.PurgeServerCacheResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		// 检查是否为节点
		_, err = this.ValidateNode(ctx)
		if err != nil {
			return nil, err
		}
	}

	if len(req.Keys) == 0 && len(req.Prefixes) == 0 {
		return &pb.PurgeServerCacheResponse{IsOk: true}, nil
	}

	var purgeResponse = &pb.PurgeServerCacheResponse{}

	var tx = this.NullTx()

	var taskType = "purge"

	var tasks = []*pb.CreateHTTPCacheTaskRequest{}
	if len(req.Keys) > 0 {
		tasks = append(tasks, &pb.CreateHTTPCacheTaskRequest{
			Type:    taskType,
			KeyType: "key",
			Keys:    req.Keys,
		})
	}
	if len(req.Prefixes) > 0 {
		tasks = append(tasks, &pb.CreateHTTPCacheTaskRequest{
			Type:    taskType,
			KeyType: "prefix",
			Keys:    req.Prefixes,
		})
	}

	var domainMap = map[string]*models.Server{} // domain name => *Server

	for _, pbTask := range tasks {
		// 创建任务
		taskId, err := models.SharedHTTPCacheTaskDAO.CreateTask(tx, 0, pbTask.Type, pbTask.KeyType, "调用PURGE API")
		if err != nil {
			return nil, err
		}

		var countKeys = 0

		for _, key := range req.Keys {
			if len(key) == 0 {
				continue
			}

			// 获取域名
			var domain = utils.ParseDomainFromKey(key)
			if len(domain) == 0 {
				continue
			}

			// 查询所在集群
			server, ok := domainMap[domain]
			if !ok {
				server, err = models.SharedServerDAO.FindEnabledServerWithDomain(tx, domain)
				if err != nil {
					return nil, err
				}
				if server == nil {
					continue
				}
				domainMap[domain] = server
			}

			var serverClusterId = int64(server.ClusterId)
			if serverClusterId == 0 {
				continue
			}

			_, err = models.SharedHTTPCacheTaskKeyDAO.CreateKey(tx, taskId, key, pbTask.Type, pbTask.KeyType, serverClusterId)
			if err != nil {
				return nil, err
			}

			countKeys++
		}

		if countKeys == 0 {
			// 如果没有有效的Key，则直接完成
			err = models.SharedHTTPCacheTaskDAO.UpdateTaskStatus(tx, taskId, true, true)
		} else {
			err = models.SharedHTTPCacheTaskDAO.UpdateTaskReady(tx, taskId)
		}
		if err != nil {
			return nil, err
		}
	}

	purgeResponse.IsOk = true

	return purgeResponse, nil
}

// FindEnabledServerTrafficLimit 查找流量限制
func (this *ServerService) FindEnabledServerTrafficLimit(ctx context.Context, req *pb.FindEnabledServerTrafficLimitRequest) (*pb.FindEnabledServerTrafficLimitResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
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
	_, err := this.ValidateAdmin(ctx)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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
			IsOn:   userPlan.IsOn,
			DayTo:  userPlan.DayTo,
			User:   nil,
			Plan: &pb.Plan{
				Id:               int64(plan.Id),
				Name:             plan.Name,
				PriceType:        plan.PriceType,
				TrafficPriceJSON: plan.TrafficPrice,
				TrafficLimitJSON: plan.TrafficLimit,
			},
		},
	}, nil
}

// ComposeServerConfig 获取服务配置
func (this *ServerService) ComposeServerConfig(ctx context.Context, req *pb.ComposeServerConfigRequest) (*pb.ComposeServerConfigResponse, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	//读取节点的所有集群
	clusterIds, err := models.SharedNodeDAO.FindEnabledNodeClusterIds(tx, nodeId)
	if err != nil {
		return nil, err
	}

	// 读取服务所在集群
	serverClusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	// 如果不在当前节点的集群中，则返回nil
	if !lists.ContainsInt64(clusterIds, serverClusterId) {
		return &pb.ComposeServerConfigResponse{ServerConfigJSON: nil}, nil
	}

	serverConfig, err := models.SharedServerDAO.ComposeServerConfigWithServerId(tx, req.ServerId, false, true)
	if err != nil {
		if err == models.ErrNotFound {
			return &pb.ComposeServerConfigResponse{ServerConfigJSON: nil}, nil
		}
		return nil, err
	}
	if serverConfig == nil {
		return &pb.ComposeServerConfigResponse{ServerConfigJSON: nil}, nil
	}

	configJSON, err := json.Marshal(serverConfig)
	if err != nil {
		return nil, err
	}
	return &pb.ComposeServerConfigResponse{ServerConfigJSON: configJSON}, nil
}

// UpdateServerUser 修改服务所属用户
func (this *ServerService) UpdateServerUser(ctx context.Context, req *pb.UpdateServerUserRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if req.ServerId <= 0 {
		return nil, errors.New("invalid serverId")
	}

	if req.UserId <= 0 {
		return nil, errors.New("invalid userId")
	}

	err = this.RunTx(func(tx *dbs.Tx) error {
		return models.SharedServerDAO.UpdateServerUserId(tx, req.ServerId, req.UserId)
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerName 修改服务名称
func (this *ServerService) UpdateServerName(ctx context.Context, req *pb.UpdateServerNameRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
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

	// 检查长度
	if len(req.Name) == 0 {
		return nil, errors.New("'name' should not be empty")
	}

	if len([]rune(req.Name)) > models.ModelServerNameMaxLength {
		return nil, errors.New("'name' too long, max length: " + types.String(models.ModelServerNameMaxLength))
	}

	err = models.SharedServerDAO.UpdateServerName(tx, req.ServerId, req.Name)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CopyServerConfig 在服务之间复制配置
func (this *ServerService) CopyServerConfig(ctx context.Context, req *pb.CopyServerConfigRequest) (*pb.RPCSuccess, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if req.ServerId <= 0 {
		return nil, errors.New("invalid 'serverId'")
	}

	// 检查权限
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	switch req.TargetType {
	case "servers":
		// 检查权限
		if len(req.TargetServerIds) == 0 {
			return this.Success()
		}
		if userId > 0 {
			for _, targetServerId := range req.TargetServerIds {
				err = models.SharedServerDAO.CheckUserServer(tx, userId, targetServerId)
				if err != nil {
					return nil, err
				}
			}
		}
		err = models.SharedServerDAO.CopyServerConfigToServers(tx, req.ServerId, req.TargetServerIds, req.ConfigCode)
		if err != nil {
			return nil, err
		}
	case "groups":
		// 检查权限
		if len(req.TargetServerGroupIds) == 0 {
			return this.Success()
		}
		if userId > 0 {
			for _, targetGroupId := range req.TargetServerGroupIds {
				err = models.SharedServerGroupDAO.CheckUserGroup(tx, userId, targetGroupId)
				if err != nil {
					return nil, err
				}
			}
		}
		err = models.SharedServerDAO.CopyServerConfigToGroups(tx, req.ServerId, req.TargetServerGroupIds, req.ConfigCode)
		if err != nil {
			return nil, err
		}
	case "cluster":
		// 检查权限
		if adminId <= 0 {
			return nil, this.PermissionError()
		}
		if req.TargetClusterId <= 0 {
			return this.Success()
		}
		err = models.SharedServerDAO.CopyServerConfigToCluster(tx, req.ServerId, req.TargetClusterId, req.ConfigCode)
		if err != nil {
			return nil, err
		}
	case "user":
		if userId == 0 {
			userId, err = models.SharedServerDAO.FindServerUserId(tx, req.ServerId)
			if err != nil {
				return nil, err
			}

			// 此时如果用户为0，则同步到未分配用户的服务
		}
		err = models.SharedServerDAO.CopyServerConfigToUser(tx, req.ServerId, req.TargetUserId, req.ConfigCode)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}
