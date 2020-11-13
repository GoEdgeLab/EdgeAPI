package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// IP名单相关服务
type IPListService struct {
}

// 创建IP列表
func (this *IPListService) CreateIPList(ctx context.Context, req *pb.CreateIPListRequest) (*pb.CreateIPListResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	listId, err := models.SharedIPListDAO.CreateIPList(req.Type, req.Name, req.Code, req.TimeoutJSON)
	if err != nil {
		return nil, err
	}
	return &pb.CreateIPListResponse{IpListId: listId}, nil
}

// 修改IP列表
func (this *IPListService) UpdateIPList(ctx context.Context, req *pb.UpdateIPListRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedIPListDAO.UpdateIPList(req.IpListId, req.Name, req.Code, req.TimeoutJSON)
	if err != nil {
		return nil, err
	}
	return rpcutils.Success()
}

// 查找IP列表
func (this *IPListService) FindEnabledIPList(ctx context.Context, req *pb.FindEnabledIPListRequest) (*pb.FindEnabledIPListResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	list, err := models.SharedIPListDAO.FindEnabledIPList(req.IpListId)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return &pb.FindEnabledIPListResponse{IpList: nil}, nil
	}
	return &pb.FindEnabledIPListResponse{IpList: &pb.IPList{
		Id:          int64(list.Id),
		IsOn:        list.IsOn == 1,
		Type:        list.Type,
		Name:        list.Name,
		Code:        list.Code,
		TimeoutJSON: []byte(list.Timeout),
	}}, nil
}
