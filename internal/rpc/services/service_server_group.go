package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ServerGroupService 服务分组相关服务
type ServerGroupService struct {
	BaseService
}

// CreateServerGroup 创建分组
func (this *ServerGroupService) CreateServerGroup(ctx context.Context, req *pb.CreateServerGroupRequest) (*pb.CreateServerGroupResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groupId, err := models.SharedServerGroupDAO.CreateGroup(tx, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateServerGroupResponse{ServerGroupId: groupId}, nil
}

// UpdateServerGroup 修改分组
func (this *ServerGroupService) UpdateServerGroup(ctx context.Context, req *pb.UpdateServerGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerGroupDAO.UpdateGroup(tx, req.ServerGroupId, req.Name)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteServerGroup 删除分组
func (this *ServerGroupService) DeleteServerGroup(ctx context.Context, req *pb.DeleteServerGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerGroupDAO.DisableServerGroup(tx, req.ServerGroupId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledServerGroups 查询所有分组
func (this *ServerGroupService) FindAllEnabledServerGroups(ctx context.Context, req *pb.FindAllEnabledServerGroupsRequest) (*pb.FindAllEnabledServerGroupsResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
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
	return &pb.FindAllEnabledServerGroupsResponse{ServerGroups: result}, nil
}

// UpdateServerGroupOrders 修改分组排序
func (this *ServerGroupService) UpdateServerGroupOrders(ctx context.Context, req *pb.UpdateServerGroupOrdersRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerGroupDAO.UpdateGroupOrders(tx, req.ServerGroupIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledServerGroup 查找单个分组信息
func (this *ServerGroupService) FindEnabledServerGroup(ctx context.Context, req *pb.FindEnabledServerGroupRequest) (*pb.FindEnabledServerGroupResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
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

	return &pb.FindEnabledServerGroupResponse{
		ServerGroup: &pb.ServerGroup{
			Id:   int64(group.Id),
			Name: group.Name,
		},
	}, nil
}
