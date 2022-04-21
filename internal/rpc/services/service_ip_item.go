package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"net"
	"time"
)

// IPItemService IP条目相关服务
type IPItemService struct {
	BaseService
}

// CreateIPItem 创建IP
func (this *IPItemService) CreateIPItem(ctx context.Context, req *pb.CreateIPItemRequest) (*pb.CreateIPItemResponse, error) {
	// 校验请求
	userType, _, userId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser, rpcutils.UserTypeNode, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	if len(req.IpFrom) == 0 {
		return nil, errors.New("'ipFrom' should not be empty")
	}

	ipFrom := net.ParseIP(req.IpFrom)
	if ipFrom == nil {
		return nil, errors.New("invalid 'ipFrom'")
	}

	if len(req.IpTo) > 0 {
		ipTo := net.ParseIP(req.IpTo)
		if ipTo == nil {
			return nil, errors.New("invalid 'ipTo'")
		}
	}

	tx := this.NullTx()

	if userType == rpcutils.UserTypeUser {
		if userId <= 0 {
			return nil, errors.New("invalid userId")
		} else {
			err = models.SharedIPListDAO.CheckUserIPList(tx, userId, req.IpListId)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(req.Type) == 0 {
		req.Type = models.IPItemTypeIPv4
	}

	// 删除以前的
	err = models.SharedIPItemDAO.DeleteOldItem(tx, req.IpListId, req.IpFrom, req.IpTo)
	if err != nil {
		return nil, err
	}

	itemId, err := models.SharedIPItemDAO.CreateIPItem(tx, req.IpListId, req.IpFrom, req.IpTo, req.ExpiredAt, req.Reason, req.Type, req.EventLevel, req.NodeId, req.ServerId, req.SourceNodeId, req.SourceServerId, req.SourceHTTPFirewallPolicyId, req.SourceHTTPFirewallRuleGroupId, req.SourceHTTPFirewallRuleSetId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateIPItemResponse{IpItemId: itemId}, nil
}

// UpdateIPItem 修改IP
func (this *IPItemService) UpdateIPItem(ctx context.Context, req *pb.UpdateIPItemRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		listId, err := models.SharedIPItemDAO.FindItemListId(tx, req.IpItemId)
		if err != nil {
			return nil, err
		}

		err = models.SharedIPListDAO.CheckUserIPList(tx, userId, listId)
		if err != nil {
			return nil, err
		}
	}

	if len(req.Type) == 0 {
		req.Type = models.IPItemTypeIPv4
	}

	err = models.SharedIPItemDAO.UpdateIPItem(tx, req.IpItemId, req.IpFrom, req.IpTo, req.ExpiredAt, req.Reason, req.Type, req.EventLevel)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteIPItem 删除IP
func (this *IPItemService) DeleteIPItem(ctx context.Context, req *pb.DeleteIPItemRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		listId, err := models.SharedIPItemDAO.FindItemListId(tx, req.IpItemId)
		if err != nil {
			return nil, err
		}

		err = models.SharedIPListDAO.CheckUserIPList(tx, userId, listId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedIPItemDAO.DisableIPItem(tx, req.IpItemId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteIPItems 批量删除IP
func (this *IPItemService) DeleteIPItems(ctx context.Context, req *pb.DeleteIPItemsRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()
	for _, itemId := range req.IpItemIds {
		err = models.SharedIPItemDAO.DisableIPItem(tx, itemId)
		if err != nil {
			return nil, err
		}
	}
	return this.Success()
}

// CountIPItemsWithListId 计算IP数量
func (this *IPItemService) CountIPItemsWithListId(ctx context.Context, req *pb.CountIPItemsWithListIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedIPListDAO.CheckUserIPList(tx, userId, req.IpListId)
		if err != nil {
			return nil, err
		}
	}

	count, err := models.SharedIPItemDAO.CountIPItemsWithListId(tx, req.IpListId, req.Keyword, req.IpFrom, req.IpTo, req.EventLevel)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListIPItemsWithListId 列出单页的IP
func (this *IPItemService) ListIPItemsWithListId(ctx context.Context, req *pb.ListIPItemsWithListIdRequest) (*pb.ListIPItemsWithListIdResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedIPListDAO.CheckUserIPList(tx, userId, req.IpListId)
		if err != nil {
			return nil, err
		}
	}

	items, err := models.SharedIPItemDAO.ListIPItemsWithListId(tx, req.IpListId, req.Keyword, req.IpFrom, req.IpTo, req.EventLevel, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.IPItem{}
	for _, item := range items {
		if len(item.Type) == 0 {
			item.Type = models.IPItemTypeIPv4
		}

		// server
		var pbSourceServer *pb.Server
		if item.SourceServerId > 0 {
			serverName, err := models.SharedServerDAO.FindEnabledServerName(tx, int64(item.SourceServerId))
			if err != nil {
				return nil, err
			}
			pbSourceServer = &pb.Server{
				Id:   int64(item.SourceServerId),
				Name: serverName,
			}
		}

		// WAF策略
		var pbSourcePolicy *pb.HTTPFirewallPolicy
		if item.SourceHTTPFirewallPolicyId > 0 {
			policy, err := models.SharedHTTPFirewallPolicyDAO.FindEnabledHTTPFirewallPolicyBasic(tx, int64(item.SourceHTTPFirewallPolicyId))
			if err != nil {
				return nil, err
			}
			if policy != nil {
				pbSourcePolicy = &pb.HTTPFirewallPolicy{
					Id:       int64(item.SourceHTTPFirewallPolicyId),
					Name:     policy.Name,
					ServerId: int64(policy.ServerId),
				}
			}
		}

		// WAF分组
		var pbSourceGroup *pb.HTTPFirewallRuleGroup
		if item.SourceHTTPFirewallRuleGroupId > 0 {
			groupName, err := models.SharedHTTPFirewallRuleGroupDAO.FindHTTPFirewallRuleGroupName(tx, int64(item.SourceHTTPFirewallRuleGroupId))
			if err != nil {
				return nil, err
			}
			pbSourceGroup = &pb.HTTPFirewallRuleGroup{
				Id:   int64(item.SourceHTTPFirewallRuleGroupId),
				Name: groupName,
			}
		}

		// WAF规则集
		var pbSourceSet *pb.HTTPFirewallRuleSet
		if item.SourceHTTPFirewallRuleSetId > 0 {
			setName, err := models.SharedHTTPFirewallRuleSetDAO.FindHTTPFirewallRuleSetName(tx, int64(item.SourceHTTPFirewallRuleSetId))
			if err != nil {
				return nil, err
			}
			pbSourceSet = &pb.HTTPFirewallRuleSet{
				Id:   int64(item.SourceHTTPFirewallRuleSetId),
				Name: setName,
			}
		}

		result = append(result, &pb.IPItem{
			Id:                            int64(item.Id),
			IpFrom:                        item.IpFrom,
			IpTo:                          item.IpTo,
			Version:                       int64(item.Version),
			CreatedAt:                     int64(item.CreatedAt),
			ExpiredAt:                     int64(item.ExpiredAt),
			Reason:                        item.Reason,
			Type:                          item.Type,
			EventLevel:                    item.EventLevel,
			NodeId:                        int64(item.NodeId),
			ServerId:                      int64(item.ServerId),
			SourceNodeId:                  int64(item.SourceNodeId),
			SourceServerId:                int64(item.SourceServerId),
			SourceHTTPFirewallPolicyId:    int64(item.SourceHTTPFirewallPolicyId),
			SourceHTTPFirewallRuleGroupId: int64(item.SourceHTTPFirewallRuleGroupId),
			SourceHTTPFirewallRuleSetId:   int64(item.SourceHTTPFirewallRuleSetId),
			SourceServer:                  pbSourceServer,
			SourceHTTPFirewallPolicy:      pbSourcePolicy,
			SourceHTTPFirewallRuleGroup:   pbSourceGroup,
			SourceHTTPFirewallRuleSet:     pbSourceSet,
			IsRead:                        item.IsRead,
		})
	}

	return &pb.ListIPItemsWithListIdResponse{IpItems: result}, nil
}

// FindEnabledIPItem 查找单个IP
func (this *IPItemService) FindEnabledIPItem(ctx context.Context, req *pb.FindEnabledIPItemRequest) (*pb.FindEnabledIPItemResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	item, err := models.SharedIPItemDAO.FindEnabledIPItem(tx, req.IpItemId)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &pb.FindEnabledIPItemResponse{IpItem: nil}, nil
	}

	if userId > 0 {
		err = models.SharedIPListDAO.CheckUserIPList(tx, userId, int64(item.ListId))
		if err != nil {
			return nil, err
		}
	}

	if len(item.Type) == 0 {
		item.Type = models.IPItemTypeIPv4
	}

	return &pb.FindEnabledIPItemResponse{IpItem: &pb.IPItem{
		Id:         int64(item.Id),
		IpFrom:     item.IpFrom,
		IpTo:       item.IpTo,
		Version:    int64(item.Version),
		CreatedAt:  int64(item.CreatedAt),
		ExpiredAt:  int64(item.ExpiredAt),
		Reason:     item.Reason,
		Type:       item.Type,
		EventLevel: item.EventLevel,
		NodeId:     int64(item.NodeId),
		ServerId:   int64(item.ServerId),
	}}, nil
}

// ListIPItemsAfterVersion 根据版本列出一组IP
func (this *IPItemService) ListIPItemsAfterVersion(ctx context.Context, req *pb.ListIPItemsAfterVersionRequest) (*pb.ListIPItemsAfterVersionResponse, error) {
	// 校验请求
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.IPItem{}
	items, err := models.SharedIPItemDAO.ListIPItemsAfterVersion(tx, req.Version, req.Size)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		// 是否已过期
		if item.ExpiredAt > 0 && int64(item.ExpiredAt) <= time.Now().Unix() {
			item.State = models.IPItemStateDisabled
		}

		if len(item.Type) == 0 {
			item.Type = models.IPItemTypeIPv4
		}

		// List类型
		list, err := models.SharedIPListDAO.FindIPListCacheable(tx, int64(item.ListId))
		if err != nil {
			return nil, err
		}
		if list == nil {
			continue
		}

		// 如果已经删除
		if list.State != models.IPListStateEnabled {
			item.State = models.IPItemStateDisabled
		}

		result = append(result, &pb.IPItem{
			Id:         int64(item.Id),
			IpFrom:     item.IpFrom,
			IpTo:       item.IpTo,
			Version:    int64(item.Version),
			CreatedAt:  int64(item.CreatedAt),
			ExpiredAt:  int64(item.ExpiredAt),
			Reason:     "", // 这里我们不需要这个数据
			ListId:     int64(item.ListId),
			IsDeleted:  item.State == 0,
			Type:       item.Type,
			EventLevel: item.EventLevel,
			ListType:   list.Type,
			IsGlobal:   list.IsPublic && list.IsGlobal,
			NodeId:     int64(item.NodeId),
			ServerId:   int64(item.ServerId),
		})
	}

	return &pb.ListIPItemsAfterVersionResponse{IpItems: result}, nil
}

// CheckIPItemStatus 检查IP状态
func (this *IPItemService) CheckIPItemStatus(ctx context.Context, req *pb.CheckIPItemStatusRequest) (*pb.CheckIPItemStatusResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// 校验IP
	ip := net.ParseIP(req.Ip)
	if len(ip) == 0 {
		return &pb.CheckIPItemStatusResponse{
			IsOk:  false,
			Error: "请输入正确的IP",
		}, nil
	}
	ipLong := utils.IP2Long(req.Ip)

	tx := this.NullTx()

	// 名单类型
	list, err := models.SharedIPListDAO.FindEnabledIPList(tx, req.IpListId, nil)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return &pb.CheckIPItemStatusResponse{
			IsOk:  false,
			Error: "IP名单不存在",
		}, nil
	}
	var isAllowed = list.Type == "white"

	// 检查IP名单
	item, err := models.SharedIPItemDAO.FindEnabledItemContainsIP(tx, req.IpListId, ipLong)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return &pb.CheckIPItemStatusResponse{
			IsOk:      true,
			Error:     "",
			IsFound:   true,
			IsAllowed: isAllowed,
			IpItem: &pb.IPItem{
				Id:         int64(item.Id),
				IpFrom:     item.IpFrom,
				IpTo:       item.IpTo,
				CreatedAt:  int64(item.CreatedAt),
				ExpiredAt:  int64(item.ExpiredAt),
				Reason:     item.Reason,
				Type:       item.Type,
				EventLevel: item.EventLevel,
			},
		}, nil
	}

	return &pb.CheckIPItemStatusResponse{
		IsOk:      true,
		Error:     "",
		IsFound:   false,
		IsAllowed: false,
		IpItem:    nil,
	}, nil
}

// ExistsEnabledIPItem 检查IP是否存在
func (this *IPItemService) ExistsEnabledIPItem(ctx context.Context, req *pb.ExistsEnabledIPItemRequest) (*pb.ExistsEnabledIPItemResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	b, err := models.SharedIPItemDAO.ExistsEnabledItem(tx, req.IpItemId)
	if err != nil {
		return nil, err
	}
	return &pb.ExistsEnabledIPItemResponse{Exists: b}, nil
}

// CountAllEnabledIPItems 计算所有IP数量
func (this *IPItemService) CountAllEnabledIPItems(ctx context.Context, req *pb.CountAllEnabledIPItemsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var listId int64 = 0
	if req.GlobalOnly {
		listId = firewallconfigs.GlobalListId
	}
	count, err := models.SharedIPItemDAO.CountAllEnabledIPItems(tx, req.Ip, listId, req.Unread, req.EventLevel, req.ListType)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListAllEnabledIPItems 搜索IP
func (this *IPItemService) ListAllEnabledIPItems(ctx context.Context, req *pb.ListAllEnabledIPItemsRequest) (*pb.ListAllEnabledIPItemsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var results = []*pb.ListAllEnabledIPItemsResponse_Result{}
	var tx = this.NullTx()
	var listId int64 = 0
	if req.GlobalOnly {
		listId = firewallconfigs.GlobalListId
	}
	items, err := models.SharedIPItemDAO.ListAllEnabledIPItems(tx, req.Ip, listId, req.Unread, req.EventLevel, req.ListType, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var cacheMap = utils.NewCacheMap()
	for _, item := range items {
		// server
		var pbSourceServer *pb.Server
		if item.SourceServerId > 0 {
			serverName, err := models.SharedServerDAO.FindEnabledServerName(tx, int64(item.SourceServerId))
			if err != nil {
				return nil, err
			}
			pbSourceServer = &pb.Server{
				Id:   int64(item.SourceServerId),
				Name: serverName,
			}
		}

		// WAF策略
		var pbSourcePolicy *pb.HTTPFirewallPolicy
		if item.SourceHTTPFirewallPolicyId > 0 {
			policy, err := models.SharedHTTPFirewallPolicyDAO.FindEnabledHTTPFirewallPolicyBasic(tx, int64(item.SourceHTTPFirewallPolicyId))
			if err != nil {
				return nil, err
			}
			if policy != nil {
				pbSourcePolicy = &pb.HTTPFirewallPolicy{
					Id:       int64(item.SourceHTTPFirewallPolicyId),
					Name:     policy.Name,
					ServerId: int64(policy.ServerId),
				}
			}
		}

		// WAF分组
		var pbSourceGroup *pb.HTTPFirewallRuleGroup
		if item.SourceHTTPFirewallRuleGroupId > 0 {
			groupName, err := models.SharedHTTPFirewallRuleGroupDAO.FindHTTPFirewallRuleGroupName(tx, int64(item.SourceHTTPFirewallRuleGroupId))
			if err != nil {
				return nil, err
			}
			pbSourceGroup = &pb.HTTPFirewallRuleGroup{
				Id:   int64(item.SourceHTTPFirewallRuleGroupId),
				Name: groupName,
			}
		}

		// WAF规则集
		var pbSourceSet *pb.HTTPFirewallRuleSet
		if item.SourceHTTPFirewallRuleSetId > 0 {
			setName, err := models.SharedHTTPFirewallRuleSetDAO.FindHTTPFirewallRuleSetName(tx, int64(item.SourceHTTPFirewallRuleSetId))
			if err != nil {
				return nil, err
			}
			pbSourceSet = &pb.HTTPFirewallRuleSet{
				Id:   int64(item.SourceHTTPFirewallRuleSetId),
				Name: setName,
			}
		}

		// 节点
		var pbSourceNode *pb.Node
		if item.SourceNodeId > 0 {
			node, err := models.SharedNodeDAO.FindEnabledBasicNode(tx, int64(item.SourceNodeId))
			if err != nil {
				return nil, err
			}
			if node != nil {
				pbSourceNode = &pb.Node{
					Id:          int64(node.Id),
					Name:        node.Name,
					NodeCluster: &pb.NodeCluster{Id: int64(node.ClusterId)},
				}
			}
		}

		var pbItem = &pb.IPItem{
			Id:                            int64(item.Id),
			IpFrom:                        item.IpFrom,
			IpTo:                          item.IpTo,
			Version:                       int64(item.Version),
			CreatedAt:                     int64(item.CreatedAt),
			ExpiredAt:                     int64(item.ExpiredAt),
			Reason:                        item.Reason,
			Type:                          item.Type,
			EventLevel:                    item.EventLevel,
			NodeId:                        int64(item.NodeId),
			ServerId:                      int64(item.ServerId),
			SourceNodeId:                  int64(item.SourceNodeId),
			SourceServerId:                int64(item.SourceServerId),
			SourceHTTPFirewallPolicyId:    int64(item.SourceHTTPFirewallPolicyId),
			SourceHTTPFirewallRuleGroupId: int64(item.SourceHTTPFirewallRuleGroupId),
			SourceHTTPFirewallRuleSetId:   int64(item.SourceHTTPFirewallRuleSetId),
			SourceServer:                  pbSourceServer,
			SourceHTTPFirewallPolicy:      pbSourcePolicy,
			SourceHTTPFirewallRuleGroup:   pbSourceGroup,
			SourceHTTPFirewallRuleSet:     pbSourceSet,
			SourceNode:                    pbSourceNode,
			IsRead:                        item.IsRead,
		}

		// 所属名单
		list, err := models.SharedIPListDAO.FindEnabledIPList(tx, int64(item.ListId), cacheMap)
		if err != nil {
			return nil, err
		}
		if list == nil {
			err = models.SharedIPItemDAO.DisableIPItem(tx, int64(item.Id))
			if err != nil {
				return nil, err
			}
			continue
		}
		var pbList = &pb.IPList{
			Id:       int64(list.Id),
			Name:     list.Name,
			Type:     list.Type,
			IsPublic: list.IsPublic,
			IsGlobal: list.IsGlobal,
		}

		// 所属服务（注意同SourceServer不同）
		var pbFirewallServer *pb.Server

		// 所属策略（注意同SourceHTTPFirewallPolicy不同）
		var pbFirewallPolicy *pb.HTTPFirewallPolicy
		if !list.IsPublic {
			policy, err := models.SharedHTTPFirewallPolicyDAO.FindEnabledFirewallPolicyWithIPListId(tx, int64(list.Id))
			if err != nil {
				return nil, err
			}
			if policy == nil {
				err = models.SharedIPItemDAO.DisableIPItem(tx, int64(item.Id))
				if err != nil {
					return nil, err
				}
				continue
			}

			pbFirewallPolicy = &pb.HTTPFirewallPolicy{
				Id:   int64(policy.Id),
				Name: policy.Name,
			}

			if policy.ServerId > 0 {
				serverName, err := models.SharedServerDAO.FindEnabledServerName(tx, int64(policy.ServerId))
				if err != nil {
					return nil, err
				}
				if len(serverName) == 0 {
					serverName = "[已删除]"
				}
				pbFirewallServer = &pb.Server{
					Id:   int64(policy.ServerId),
					Name: serverName,
				}
			}
		}

		results = append(results, &pb.ListAllEnabledIPItemsResponse_Result{
			IpList:             pbList,
			IpItem:             pbItem,
			Server:             pbFirewallServer,
			HttpFirewallPolicy: pbFirewallPolicy,
		})
	}

	return &pb.ListAllEnabledIPItemsResponse{Results: results}, nil
}

// UpdateIPItemsRead 设置所有为已读
func (this *IPItemService) UpdateIPItemsRead(ctx context.Context, req *pb.UpdateIPItemsReadRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedIPItemDAO.UpdateItemsRead(tx)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
