package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 节点分组相关服务
type NodeGroupService struct {
}

// 创建分组
func (this *NodeGroupService) CreateNodeGroup(ctx context.Context, req *pb.CreateNodeGroupRequest) (*pb.CreateNodeGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	groupId, err := models.SharedNodeGroupDAO.CreateNodeGroup(req.ClusterId, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodeGroupResponse{GroupId: groupId}, nil
}

// 修改分组
func (this *NodeGroupService) UpdateNodeGroup(ctx context.Context, req *pb.UpdateNodeGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeGroupDAO.UpdateNodeGroup(req.GroupId, req.Name)
	if err != nil {
		return nil, err
	}

	return rpcutils.Success()
}

// 删除分组
func (this *NodeGroupService) DeleteNodeGroup(ctx context.Context, req *pb.DeleteNodeGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	_, err = models.SharedNodeGroupDAO.DisableNodeGroup(req.GroupId)
	if err != nil {
		return nil, err
	}

	return rpcutils.Success()
}

// 查询所有分组
func (this *NodeGroupService) FindAllEnabledNodeGroupsWithClusterId(ctx context.Context, req *pb.FindAllEnabledNodeGroupsWithClusterIdRequest) (*pb.FindAllEnabledNodeGroupsWithClusterIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	groups, err := models.SharedNodeGroupDAO.FindAllEnabledGroupsWithClusterId(req.ClusterId)
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
	return &pb.FindAllEnabledNodeGroupsWithClusterIdResponse{Groups: result}, nil
}

// 修改分组排序
func (this *NodeGroupService) UpdateNodeGroupOrders(ctx context.Context, req *pb.UpdateNodeGroupOrdersRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeGroupDAO.UpdateGroupOrders(req.GroupIds)
	if err != nil {
		return nil, err
	}
	return rpcutils.Success()
}

// 查找单个分组信息
func (this *NodeGroupService) FindEnabledNodeGroup(ctx context.Context, req *pb.FindEnabledNodeGroupRequest) (*pb.FindEnabledNodeGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	group, err := models.SharedNodeGroupDAO.FindEnabledNodeGroup(req.GroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindEnabledNodeGroupResponse{
			Group: nil,
		}, nil
	}

	return &pb.FindEnabledNodeGroupResponse{
		Group: &pb.NodeGroup{
			Id:   int64(group.Id),
			Name: group.Name,
		},
	}, nil
}
