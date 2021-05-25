package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NodeGroupService 节点分组相关服务
type NodeGroupService struct {
	BaseService
}

// CreateNodeGroup 创建分组
func (this *NodeGroupService) CreateNodeGroup(ctx context.Context, req *pb.CreateNodeGroupRequest) (*pb.CreateNodeGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groupId, err := models.SharedNodeGroupDAO.CreateNodeGroup(tx, req.NodeClusterId, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodeGroupResponse{NodeGroupId: groupId}, nil
}

// UpdateNodeGroup 修改分组
func (this *NodeGroupService) UpdateNodeGroup(ctx context.Context, req *pb.UpdateNodeGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeGroupDAO.UpdateNodeGroup(tx, req.NodeGroupId, req.Name)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteNodeGroup 删除分组
func (this *NodeGroupService) DeleteNodeGroup(ctx context.Context, req *pb.DeleteNodeGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	_, err = models.SharedNodeGroupDAO.DisableNodeGroup(tx, req.NodeGroupId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledNodeGroupsWithNodeClusterId 查询所有分组
func (this *NodeGroupService) FindAllEnabledNodeGroupsWithNodeClusterId(ctx context.Context, req *pb.FindAllEnabledNodeGroupsWithNodeClusterIdRequest) (*pb.FindAllEnabledNodeGroupsWithNodeClusterIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groups, err := models.SharedNodeGroupDAO.FindAllEnabledGroupsWithClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result := []*pb.NodeGroup{}
	for _, group := range groups {
		result = append(result, &pb.NodeGroup{
			Id:   int64(group.Id),
			Name: group.Name,
		})
	}
	return &pb.FindAllEnabledNodeGroupsWithNodeClusterIdResponse{NodeGroups: result}, nil
}

// UpdateNodeGroupOrders 修改分组排序
func (this *NodeGroupService) UpdateNodeGroupOrders(ctx context.Context, req *pb.UpdateNodeGroupOrdersRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedNodeGroupDAO.UpdateGroupOrders(tx, req.NodeGroupIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledNodeGroup 查找单个分组信息
func (this *NodeGroupService) FindEnabledNodeGroup(ctx context.Context, req *pb.FindEnabledNodeGroupRequest) (*pb.FindEnabledNodeGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	group, err := models.SharedNodeGroupDAO.FindEnabledNodeGroup(tx, req.NodeGroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindEnabledNodeGroupResponse{
			NodeGroup: nil,
		}, nil
	}

	return &pb.FindEnabledNodeGroupResponse{
		NodeGroup: &pb.NodeGroup{
			Id:   int64(group.Id),
			Name: group.Name,
		},
	}, nil
}
