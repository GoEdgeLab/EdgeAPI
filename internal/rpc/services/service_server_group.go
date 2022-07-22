package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

// ServerGroupService 服务分组相关服务
type ServerGroupService struct {
	BaseService
}

// CreateServerGroup 创建分组
func (this *ServerGroupService) CreateServerGroup(ctx context.Context, req *pb.CreateServerGroupRequest) (*pb.CreateServerGroupResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groupId, err := models.SharedServerGroupDAO.CreateGroup(tx, req.Name, userId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateServerGroupResponse{ServerGroupId: groupId}, nil
}

// UpdateServerGroup 修改分组
func (this *ServerGroupService) UpdateServerGroup(ctx context.Context, req *pb.UpdateServerGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = models.SharedServerGroupDAO.CheckUserGroup(tx, userId, req.ServerGroupId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedServerGroupDAO.UpdateGroup(tx, req.ServerGroupId, req.Name)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteServerGroup 删除分组
func (this *ServerGroupService) DeleteServerGroup(ctx context.Context, req *pb.DeleteServerGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查用户权限
	if userId > 0 {
		err = models.SharedServerGroupDAO.CheckUserGroup(tx, userId, req.ServerGroupId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedServerGroupDAO.DisableServerGroup(tx, req.ServerGroupId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledServerGroups 查询所有分组
func (this *ServerGroupService) FindAllEnabledServerGroups(ctx context.Context, req *pb.FindAllEnabledServerGroupsRequest) (*pb.FindAllEnabledServerGroupsResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groups, err := models.SharedServerGroupDAO.FindAllEnabledGroups(tx, userId)
	if err != nil {
		return nil, err
	}
	result := []*pb.ServerGroup{}
	for _, group := range groups {
		result = append(result, &pb.ServerGroup{
			Id:   int64(group.Id),
			Name: group.Name,
		})
	}
	return &pb.FindAllEnabledServerGroupsResponse{ServerGroups: result}, nil
}

// UpdateServerGroupOrders 修改分组排序
func (this *ServerGroupService) UpdateServerGroupOrders(ctx context.Context, req *pb.UpdateServerGroupOrdersRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerGroupDAO.UpdateGroupOrders(tx, req.ServerGroupIds, userId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledServerGroup 查找单个分组信息
func (this *ServerGroupService) FindEnabledServerGroup(ctx context.Context, req *pb.FindEnabledServerGroupRequest) (*pb.FindEnabledServerGroupResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(tx, req.ServerGroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindEnabledServerGroupResponse{
			ServerGroup: nil,
		}, nil
	}

	// 检查用户权限
	if userId > 0 && int64(group.UserId) != userId {
		return &pb.FindEnabledServerGroupResponse{
			ServerGroup: nil,
		}, nil
	}

	return &pb.FindEnabledServerGroupResponse{
		ServerGroup: &pb.ServerGroup{
			Id:   int64(group.Id),
			Name: group.Name,
		},
	}, nil
}

// FindAndInitServerGroupHTTPReverseProxyConfig 查找HTTP反向代理设置
func (this *ServerGroupService) FindAndInitServerGroupHTTPReverseProxyConfig(ctx context.Context, req *pb.FindAndInitServerGroupHTTPReverseProxyConfigRequest) (*pb.FindAndInitServerGroupHTTPReverseProxyConfigResponse, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	reverseProxyRef, err := models.SharedServerGroupDAO.FindHTTPReverseProxyRef(tx, req.ServerGroupId)
	if err != nil {
		return nil, err
	}

	if reverseProxyRef == nil {
		reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, 0, nil, nil, nil)
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
		err = models.SharedServerGroupDAO.UpdateHTTPReverseProxy(tx, req.ServerGroupId, refJSON)
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

	return &pb.FindAndInitServerGroupHTTPReverseProxyConfigResponse{ReverseProxyJSON: configJSON, ReverseProxyRefJSON: refJSON}, nil
}

// FindAndInitServerGroupTCPReverseProxyConfig 查找反向代理设置
func (this *ServerGroupService) FindAndInitServerGroupTCPReverseProxyConfig(ctx context.Context, req *pb.FindAndInitServerGroupTCPReverseProxyConfigRequest) (*pb.FindAndInitServerGroupTCPReverseProxyConfigResponse, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	reverseProxyRef, err := models.SharedServerGroupDAO.FindTCPReverseProxyRef(tx, req.ServerGroupId)
	if err != nil {
		return nil, err
	}

	if reverseProxyRef == nil {
		reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, 0, nil, nil, nil)
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
		err = models.SharedServerGroupDAO.UpdateTCPReverseProxy(tx, req.ServerGroupId, refJSON)
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

	return &pb.FindAndInitServerGroupTCPReverseProxyConfigResponse{ReverseProxyJSON: configJSON, ReverseProxyRefJSON: refJSON}, nil
}

// FindAndInitServerGroupUDPReverseProxyConfig 查找反向代理设置
func (this *ServerGroupService) FindAndInitServerGroupUDPReverseProxyConfig(ctx context.Context, req *pb.FindAndInitServerGroupUDPReverseProxyConfigRequest) (*pb.FindAndInitServerGroupUDPReverseProxyConfigResponse, error) {
	// 校验请求
	adminId, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	reverseProxyRef, err := models.SharedServerGroupDAO.FindUDPReverseProxyRef(tx, req.ServerGroupId)
	if err != nil {
		return nil, err
	}

	if reverseProxyRef == nil {
		reverseProxyId, err := models.SharedReverseProxyDAO.CreateReverseProxy(tx, adminId, 0, nil, nil, nil)
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
		err = models.SharedServerGroupDAO.UpdateUDPReverseProxy(tx, req.ServerGroupId, refJSON)
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

	return &pb.FindAndInitServerGroupUDPReverseProxyConfigResponse{ReverseProxyJSON: configJSON, ReverseProxyRefJSON: refJSON}, nil
}

// UpdateServerGroupHTTPReverseProxy 修改服务的反向代理设置
func (this *ServerGroupService) UpdateServerGroupHTTPReverseProxy(ctx context.Context, req *pb.UpdateServerGroupHTTPReverseProxyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerGroupDAO.UpdateHTTPReverseProxy(tx, req.ServerGroupId, req.ReverseProxyJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerGroupTCPReverseProxy 修改服务的反向代理设置
func (this *ServerGroupService) UpdateServerGroupTCPReverseProxy(ctx context.Context, req *pb.UpdateServerGroupTCPReverseProxyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerGroupDAO.UpdateTCPReverseProxy(tx, req.ServerGroupId, req.ReverseProxyJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateServerGroupUDPReverseProxy 修改服务的反向代理设置
func (this *ServerGroupService) UpdateServerGroupUDPReverseProxy(ctx context.Context, req *pb.UpdateServerGroupUDPReverseProxyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 修改配置
	err = models.SharedServerGroupDAO.UpdateUDPReverseProxy(tx, req.ServerGroupId, req.ReverseProxyJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledServerGroupConfigInfo 取得分组的配置概要信息
func (this *ServerGroupService) FindEnabledServerGroupConfigInfo(ctx context.Context, req *pb.FindEnabledServerGroupConfigInfoRequest) (*pb.FindEnabledServerGroupConfigInfoResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查用户权限
	if userId > 0 {
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}

		if req.ServerGroupId > 0 {
			err = models.SharedServerGroupDAO.CheckUserGroup(tx, userId, req.ServerGroupId)
			if err != nil {
				return nil, err
			}
		}
	}

	var group *models.ServerGroup
	if req.ServerGroupId > 0 {
		group, err = models.SharedServerGroupDAO.FindEnabledServerGroup(tx, req.ServerGroupId)
		if err != nil {
			return nil, err
		}
	} else if req.ServerId > 0 {
		groupIds, err := models.SharedServerDAO.FindServerGroupIds(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
		if len(groupIds) > 0 {
			for _, groupId := range groupIds {
				group, err = models.SharedServerGroupDAO.FindEnabledServerGroup(tx, groupId)
				if err != nil {
					return nil, err
				}
				if group != nil {
					break
				}
			}
		}
	}

	if group == nil {
		return &pb.FindEnabledServerGroupConfigInfoResponse{}, nil
	}

	var result = &pb.FindEnabledServerGroupConfigInfoResponse{
		ServerGroupId: int64(group.Id),
	}

	// http
	if len(group.HttpReverseProxy) > 0 {
		var ref = &serverconfigs.ReverseProxyRef{}
		err = json.Unmarshal(group.HttpReverseProxy, ref)
		if err != nil {
			return nil, err
		}
		result.HasHTTPReverseProxy = ref.IsPrior
	}

	// tcp
	if len(group.TcpReverseProxy) > 0 {
		var ref = &serverconfigs.ReverseProxyRef{}
		err = json.Unmarshal(group.TcpReverseProxy, ref)
		if err != nil {
			return nil, err
		}
		result.HasTCPReverseProxy = ref.IsPrior
	}

	// udp
	if len(group.UdpReverseProxy) > 0 {
		var ref = &serverconfigs.ReverseProxyRef{}
		err = json.Unmarshal(group.UdpReverseProxy, ref)
		if err != nil {
			return nil, err
		}
		result.HasUDPReverseProxy = ref.IsPrior
	}

	config, err := models.SharedServerGroupDAO.ComposeGroupConfig(tx, int64(group.Id), nil)
	if err != nil {
		return nil, err
	}
	if config != nil {
		var webConfig = config.Web
		if webConfig != nil {
			result.HasRootConfig = webConfig != nil && webConfig.Root != nil && webConfig.Root.IsPrior
			result.HasWAFConfig = webConfig != nil && webConfig.FirewallRef != nil && webConfig.FirewallRef.IsPrior
			result.HasCacheConfig = webConfig != nil && webConfig.Cache != nil && webConfig.Cache.IsPrior
			result.HasCharsetConfig = webConfig != nil && webConfig.Charset != nil && webConfig.Charset.IsPrior
			result.HasAccessLogConfig = webConfig != nil && webConfig.AccessLogRef != nil && webConfig.AccessLogRef.IsPrior
			result.HasStatConfig = webConfig != nil && webConfig.StatRef != nil && webConfig.StatRef.IsPrior
			result.HasCompressionConfig = webConfig != nil && webConfig.Compression != nil && webConfig.Compression.IsPrior
			result.HasWebsocketConfig = webConfig != nil && webConfig.WebsocketRef != nil && webConfig.WebsocketRef.IsPrior
			result.HasRequestHeadersConfig = webConfig != nil && webConfig.RequestHeaderPolicyRef != nil && webConfig.RequestHeaderPolicyRef.IsPrior
			result.HasResponseHeadersConfig = webConfig != nil && webConfig.ResponseHeaderPolicyRef != nil && webConfig.ResponseHeaderPolicyRef.IsPrior
			result.HasWebPConfig = webConfig != nil && webConfig.WebP != nil && webConfig.WebP.IsPrior
			result.HasRemoteAddrConfig = webConfig != nil && webConfig.RemoteAddr != nil && webConfig.RemoteAddr.IsPrior
			result.HasPagesConfig = webConfig != nil && (len(webConfig.Pages) > 0 || (webConfig.Shutdown != nil && webConfig.Shutdown.IsOn))
			result.HasRequestLimitConfig = webConfig != nil && webConfig.RequestLimit != nil && webConfig.RequestLimit.IsPrior
		}
	}

	return result, nil
}

// FindAndInitServerGroupWebConfig 初始化Web设置
func (this *ServerGroupService) FindAndInitServerGroupWebConfig(ctx context.Context, req *pb.FindAndInitServerGroupWebConfigRequest) (*pb.FindAndInitServerGroupWebConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	webId, err := models.SharedServerGroupDAO.FindGroupWebId(tx, req.ServerGroupId)
	if err != nil {
		return nil, err
	}

	if webId == 0 {
		webId, err = models.SharedServerGroupDAO.InitGroupWeb(tx, req.ServerGroupId)
		if err != nil {
			return nil, err
		}
	}

	webConfig, err := models.SharedHTTPWebDAO.ComposeWebConfig(tx, webId, nil)
	if err != nil {
		return nil, err
	}
	webConfigJSON, err := json.Marshal(webConfig)
	if err != nil {
		return nil, err
	}
	return &pb.FindAndInitServerGroupWebConfigResponse{WebJSON: webConfigJSON}, nil
}
