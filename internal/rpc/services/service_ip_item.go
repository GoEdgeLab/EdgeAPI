package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"net"
)

// IPItemService IP条目相关服务
type IPItemService struct {
	BaseService
}

// CreateIPItem 创建IP
func (this *IPItemService) CreateIPItem(ctx context.Context, req *pb.CreateIPItemRequest) (*pb.CreateIPItemResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
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

	if userId > 0 {
		err = models.SharedIPListDAO.CheckUserIPList(tx, userId, req.IpListId)
		if err != nil {
			return nil, err
		}
	}

	if len(req.Type) == 0 {
		req.Type = models.IPItemTypeIPv4
	}

	itemId, err := models.SharedIPItemDAO.CreateIPItem(tx, req.IpListId, req.IpFrom, req.IpTo, req.ExpiredAt, req.Reason, req.Type, req.EventLevel)
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

	count, err := models.SharedIPItemDAO.CountIPItemsWithListId(tx, req.IpListId)
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

	items, err := models.SharedIPItemDAO.ListIPItemsWithListId(tx, req.IpListId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.IPItem{}
	for _, item := range items {
		if len(item.Type) == 0 {
			item.Type = models.IPItemTypeIPv4
		}

		result = append(result, &pb.IPItem{
			Id:         int64(item.Id),
			IpFrom:     item.IpFrom,
			IpTo:       item.IpTo,
			Version:    int64(item.Version),
			ExpiredAt:  int64(item.ExpiredAt),
			Reason:     item.Reason,
			Type:       item.Type,
			EventLevel: item.EventLevel,
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
		ExpiredAt:  int64(item.ExpiredAt),
		Reason:     item.Reason,
		Type:       item.Type,
		EventLevel: item.EventLevel,
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
		if len(item.Type) == 0 {
			item.Type = models.IPItemTypeIPv4
		}

		// List类型
		listType, err := models.SharedIPListDAO.FindIPListTypeCacheable(tx, int64(item.ListId))
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.IPItem{
			Id:         int64(item.Id),
			IpFrom:     item.IpFrom,
			IpTo:       item.IpTo,
			Version:    int64(item.Version),
			ExpiredAt:  int64(item.ExpiredAt),
			Reason:     "", // 这里我们不需要这个数据
			ListId:     int64(item.ListId),
			IsDeleted:  item.State == 0,
			Type:       item.Type,
			EventLevel: item.EventLevel,
			ListType:   listType,
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
	list, err := models.SharedIPListDAO.FindEnabledIPList(tx, req.IpListId)
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
