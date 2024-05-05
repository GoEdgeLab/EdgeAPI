package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/rands"
)

// IPListService IP名单相关服务
type IPListService struct {
	BaseService
}

// CreateIPList 创建IP列表
func (this *IPListService) CreateIPList(ctx context.Context, req *pb.CreateIPListRequest) (*pb.CreateIPListResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 修正默认的代号
	if req.Code == "white" || req.Code == "black" {
		req.Code = req.Code + "-" + rands.HexString(8)
	}

	// 检查用户相关信息
	if userId > 0 {
		// 检查网站ID
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}
	}

	// 检查代号
	if len(req.Code) > 0 {
		if !models.SharedIPListDAO.ValidateIPListCode(req.Code) {
			return nil, errors.New("invalid 'code' format")
		}

		oldListId, findErr := models.SharedIPListDAO.FindIPListIdWithCode(tx, req.Code)
		if findErr != nil {
			return nil, findErr
		}
		if oldListId > 0 {
			return nil, errors.New("the code '" + req.Code + "' has been used")
		}
	}

	listId, err := models.SharedIPListDAO.CreateIPList(tx, userId, req.ServerId, req.Type, req.Name, req.Code, req.TimeoutJSON, req.Description, req.IsPublic, req.IsGlobal)
	if err != nil {
		return nil, err
	}
	return &pb.CreateIPListResponse{IpListId: listId}, nil
}

// UpdateIPList 修改IP列表
func (this *IPListService) UpdateIPList(ctx context.Context, req *pb.UpdateIPListRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查代号
	if len(req.Code) > 0 {
		if !models.SharedIPListDAO.ValidateIPListCode(req.Code) {
			return nil, errors.New("invalid 'code' format")
		}

		oldListId, findErr := models.SharedIPListDAO.FindIPListIdWithCode(tx, req.Code)
		if findErr != nil {
			return nil, findErr
		}
		if oldListId > 0 && oldListId != req.IpListId {
			return nil, errors.New("the code '" + req.Code + "' has been used")
		}
	}

	err = models.SharedIPListDAO.UpdateIPList(tx, req.IpListId, req.Name, req.Code, req.TimeoutJSON, req.Description)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledIPList 查找IP列表
func (this *IPListService) FindEnabledIPList(ctx context.Context, req *pb.FindEnabledIPListRequest) (*pb.FindEnabledIPListResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		// 检查用户所属名单
		if req.IpListId != firewallconfigs.GlobalListId {
			err = models.SharedIPListDAO.CheckUserIPList(tx, userId, req.IpListId)
			if err != nil {
				return nil, err
			}
		}
	}

	list, err := models.SharedIPListDAO.FindEnabledIPList(tx, req.IpListId, nil)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return &pb.FindEnabledIPListResponse{IpList: nil}, nil
	}
	return &pb.FindEnabledIPListResponse{IpList: &pb.IPList{
		Id:          int64(list.Id),
		IsOn:        list.IsOn,
		Type:        list.Type,
		Name:        list.Name,
		Code:        list.Code,
		TimeoutJSON: list.Timeout,
		Description: list.Description,
		IsGlobal:    list.IsGlobal,
	}}, nil
}

// CountAllEnabledIPLists 计算名单数量
func (this *IPListService) CountAllEnabledIPLists(ctx context.Context, req *pb.CountAllEnabledIPListsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedIPListDAO.CountAllEnabledIPLists(tx, req.Type, req.IsPublic, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledIPLists 列出单页名单
func (this *IPListService) ListEnabledIPLists(ctx context.Context, req *pb.ListEnabledIPListsRequest) (*pb.ListEnabledIPListsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	ipLists, err := models.SharedIPListDAO.ListEnabledIPLists(tx, req.Type, req.IsPublic, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbLists []*pb.IPList
	for _, list := range ipLists {
		pbLists = append(pbLists, &pb.IPList{
			Id:          int64(list.Id),
			IsOn:        list.IsOn,
			Type:        list.Type,
			Name:        list.Name,
			Code:        list.Code,
			TimeoutJSON: list.Timeout,
			IsPublic:    list.IsPublic,
			Description: list.Description,
			IsGlobal:    list.IsGlobal,
		})
	}
	return &pb.ListEnabledIPListsResponse{IpLists: pbLists}, nil
}

// DeleteIPList 删除IP名单
func (this *IPListService) DeleteIPList(ctx context.Context, req *pb.DeleteIPListRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedIPListDAO.DisableIPList(tx, req.IpListId)
	if err != nil {
		return nil, err
	}

	// 删除所有IP
	err = models.SharedIPItemDAO.DisableIPItemsWithListId(tx, req.IpListId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ExistsEnabledIPList 检查IPList是否存在
func (this *IPListService) ExistsEnabledIPList(ctx context.Context, req *pb.ExistsEnabledIPListRequest) (*pb.ExistsEnabledIPListResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	b, err := models.SharedIPListDAO.ExistsEnabledIPList(tx, req.IpListId)
	if err != nil {
		return nil, err
	}
	return &pb.ExistsEnabledIPListResponse{Exists: b}, nil
}

// FindEnabledIPListContainsIP 根据IP来搜索IP名单
func (this *IPListService) FindEnabledIPListContainsIP(ctx context.Context, req *pb.FindEnabledIPListContainsIPRequest) (*pb.FindEnabledIPListContainsIPResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	items, err := models.SharedIPItemDAO.FindEnabledItemsWithIP(tx, req.Ip)
	if err != nil {
		return nil, err
	}

	var pbLists = []*pb.IPList{}
	var listIds = []int64{}
	var cacheMap = utils.NewCacheMap()
	for _, item := range items {
		if lists.ContainsInt64(listIds, int64(item.ListId)) {
			continue
		}

		list, err := models.SharedIPListDAO.FindEnabledIPList(tx, int64(item.ListId), cacheMap)
		if err != nil {
			return nil, err
		}
		if list == nil {
			continue
		}
		if !list.IsPublic {
			continue
		}
		pbLists = append(pbLists, &pb.IPList{
			Id:          int64(list.Id),
			IsOn:        list.IsOn,
			Type:        list.Type,
			Name:        list.Name,
			Code:        list.Code,
			IsPublic:    list.IsPublic,
			IsGlobal:    list.IsGlobal,
			Description: "",
		})

		listIds = append(listIds, int64(item.ListId))
	}
	return &pb.FindEnabledIPListContainsIPResponse{IpLists: pbLists}, nil
}

// FindServerIdWithIPListId 查找IP名单对应的网站ID
func (this *IPListService) FindServerIdWithIPListId(ctx context.Context, req *pb.FindServerIdWithIPListIdRequest) (*pb.FindServerIdWithIPListIdResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	serverId, err := models.SharedIPListDAO.FindServerIdWithListId(tx, req.IpListId)
	if err != nil {
		return nil, err
	}

	// check user
	if serverId > 0 && userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, serverId)
		if err != nil {
			return nil, err
		}
	}

	return &pb.FindServerIdWithIPListIdResponse{
		ServerId: serverId,
	}, nil
}

// FindIPListIdWithCode 根据IP名单代号获取IP名单ID
func (this *IPListService) FindIPListIdWithCode(ctx context.Context, req *pb.FindIPListIdWithCodeRequest) (*pb.FindIPListIdWithCodeResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if len(req.Code) == 0 {
		return nil, errors.New("require 'code'")
	}

	var tx = this.NullTx()
	listId, err := models.SharedIPListDAO.FindIPListIdWithCode(tx, req.Code)
	if err != nil {
		return nil, err
	}

	if listId > 0 {
		if userId > 0 {
			err = models.SharedIPListDAO.CheckUserIPList(tx, userId, listId)
			if err != nil {
				return nil, err
			}
		}
	}

	return &pb.FindIPListIdWithCodeResponse{
		IpListId: listId,
	}, nil
}
