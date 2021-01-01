package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 服务分组相关服务
type ServerGroupService struct {
	BaseService
}

// 创建分组
func (this *ServerGroupService) CreateServerGroup(ctx context.Context, req *pb.CreateServerGroupRequest) (*pb.CreateServerGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groupId, err := models.SharedServerGroupDAO.CreateGroup(tx, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateServerGroupResponse{GroupId: groupId}, nil
}

// 修改分组
func (this *ServerGroupService) UpdateServerGroup(ctx context.Context, req *pb.UpdateServerGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerGroupDAO.UpdateGroup(tx, req.GroupId, req.Name)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 删除分组
func (this *ServerGroupService) DeleteServerGroup(ctx context.Context, req *pb.DeleteServerGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerGroupDAO.DisableServerGroup(tx, req.GroupId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查询所有分组
func (this *ServerGroupService) FindAllEnabledServerGroups(ctx context.Context, req *pb.FindAllEnabledServerGroupsRequest) (*pb.FindAllEnabledServerGroupsResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groups, err := models.SharedServerGroupDAO.FindAllEnabledGroups(tx)
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
	return &pb.FindAllEnabledServerGroupsResponse{Groups: result}, nil
}

// 修改分组排序
func (this *ServerGroupService) UpdateServerGroupOrders(ctx context.Context, req *pb.UpdateServerGroupOrdersRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerGroupDAO.UpdateGroupOrders(tx, req.GroupIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 查找单个分组信息
func (this *ServerGroupService) FindEnabledServerGroup(ctx context.Context, req *pb.FindEnabledServerGroupRequest) (*pb.FindEnabledServerGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	group, err := models.SharedServerGroupDAO.FindEnabledServerGroup(tx, req.GroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindEnabledServerGroupResponse{
			Group: nil,
		}, nil
	}

	return &pb.FindEnabledServerGroupResponse{
		Group: &pb.ServerGroup{
			Id:   int64(group.Id),
			Name: group.Name,
		},
	}, nil
}
