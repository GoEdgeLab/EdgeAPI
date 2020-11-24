package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// IP条目相关服务
type IPItemService struct {
	BaseService
}

// 创建IP
func (this *IPItemService) CreateIPItem(ctx context.Context, req *pb.CreateIPItemRequest) (*pb.CreateIPItemResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	itemId, err := models.SharedIPItemDAO.CreateIPItem(req.IpListId, req.IpFrom, req.IpTo, req.ExpiredAt, req.Reason)
	if err != nil {
		return nil, err
	}



	return &pb.CreateIPItemResponse{IpItemId: itemId}, nil
}

// 修改IP
func (this *IPItemService) UpdateIPItem(ctx context.Context, req *pb.UpdateIPItemRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedIPItemDAO.UpdateIPItem(req.IpItemId, req.IpFrom, req.IpTo, req.ExpiredAt, req.Reason)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 删除IP
func (this *IPItemService) DeleteIPItem(ctx context.Context, req *pb.DeleteIPItemRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedIPItemDAO.DisableIPItem(req.IpItemId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 计算IP数量
func (this *IPItemService) CountIPItemsWithListId(ctx context.Context, req *pb.CountIPItemsWithListIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedIPItemDAO.CountIPItemsWithListId(req.IpListId)
	if err != nil {
		return nil, err
	}
	return &pb.RPCCountResponse{Count: count}, nil
}

// 列出单页的IP
func (this *IPItemService) ListIPItemsWithListId(ctx context.Context, req *pb.ListIPItemsWithListIdRequest) (*pb.ListIPItemsWithListIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	items, err := models.SharedIPItemDAO.ListIPItemsWithListId(req.IpListId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.IPItem{}
	for _, item := range items {
		result = append(result, &pb.IPItem{
			Id:        int64(item.Id),
			IpFrom:    item.IpFrom,
			IpTo:      item.IpTo,
			Version:   int64(item.Version),
			ExpiredAt: int64(item.ExpiredAt),
			Reason:    item.Reason,
		})
	}

	return &pb.ListIPItemsWithListIdResponse{IpItems: result}, nil
}

// 查找单个IP
func (this *IPItemService) FindEnabledIPItem(ctx context.Context, req *pb.FindEnabledIPItemRequest) (*pb.FindEnabledIPItemResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	item, err := models.SharedIPItemDAO.FindEnabledIPItem(req.IpItemId)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &pb.FindEnabledIPItemResponse{IpItem: nil}, nil
	}
	return &pb.FindEnabledIPItemResponse{IpItem: &pb.IPItem{
		Id:        int64(item.Id),
		IpFrom:    item.IpFrom,
		IpTo:      item.IpTo,
		Version:   int64(item.Version),
		ExpiredAt: int64(item.ExpiredAt),
		Reason:    item.Reason,
	}}, nil
}

// 根据版本列出一组IP
func (this *IPItemService) ListIPItemsAfterVersion(ctx context.Context, req *pb.ListIPItemsAfterVersionRequest) (*pb.ListIPItemsAfterVersionResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	result := []*pb.IPItem{}
	items, err := models.SharedIPItemDAO.ListIPItemsAfterVersion(req.Version, req.Size)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		result = append(result, &pb.IPItem{
			Id:        int64(item.Id),
			IpFrom:    item.IpFrom,
			IpTo:      item.IpTo,
			Version:   int64(item.Version),
			ExpiredAt: int64(item.ExpiredAt),
			Reason:    "", // 这里我们不需要这个数据
			ListId:    int64(item.ListId),
			IsDeleted: item.State == 0,
		})
	}

	return &pb.ListIPItemsAfterVersionResponse{IpItems: result}, nil
}
