package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"google.golang.org/grpc/metadata"
)

type MonitorNodeService struct {
	BaseService
}

// 创建监控节点
func (this *MonitorNodeService) CreateMonitorNode(ctx context.Context, req *pb.CreateMonitorNodeRequest) (*pb.CreateMonitorNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodeId, err := models.SharedMonitorNodeDAO.CreateMonitorNode(tx, req.Name, req.Description, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMonitorNodeResponse{NodeId: nodeId}, nil
}

// 修改监控节点
func (this *MonitorNodeService) UpdateMonitorNode(ctx context.Context, req *pb.UpdateMonitorNodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedMonitorNodeDAO.UpdateMonitorNode(tx, req.NodeId, req.Name, req.Description, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 删除监控节点
func (this *MonitorNodeService) DeleteMonitorNode(ctx context.Context, req *pb.DeleteMonitorNodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedMonitorNodeDAO.DisableMonitorNode(tx, req.NodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 列出所有可用监控节点
func (this *MonitorNodeService) FindAllEnabledMonitorNodes(ctx context.Context, req *pb.FindAllEnabledMonitorNodesRequest) (*pb.FindAllEnabledMonitorNodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedMonitorNodeDAO.FindAllEnabledMonitorNodes(tx)
	if err != nil {
		return nil, err
	}

	result := []*pb.MonitorNode{}
	for _, node := range nodes {
		result = append(result, &pb.MonitorNode{
			Id:          int64(node.Id),
			IsOn:        node.IsOn == 1,
			UniqueId:    node.UniqueId,
			Secret:      node.Secret,
			Name:        node.Name,
			Description: node.Description,
		})
	}

	return &pb.FindAllEnabledMonitorNodesResponse{Nodes: result}, nil
}

// 计算监控节点数量
func (this *MonitorNodeService) CountAllEnabledMonitorNodes(ctx context.Context, req *pb.CountAllEnabledMonitorNodesRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedMonitorNodeDAO.CountAllEnabledMonitorNodes(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// 列出单页的监控节点
func (this *MonitorNodeService) ListEnabledMonitorNodes(ctx context.Context, req *pb.ListEnabledMonitorNodesRequest) (*pb.ListEnabledMonitorNodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedMonitorNodeDAO.ListEnabledMonitorNodes(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.MonitorNode{}
	for _, node := range nodes {
		result = append(result, &pb.MonitorNode{
			Id:          int64(node.Id),
			IsOn:        node.IsOn == 1,
			UniqueId:    node.UniqueId,
			Secret:      node.Secret,
			Name:        node.Name,
			Description: node.Description,
			StatusJSON:  []byte(node.Status),
		})
	}

	return &pb.ListEnabledMonitorNodesResponse{Nodes: result}, nil
}

// 根据ID查找节点
func (this *MonitorNodeService) FindEnabledMonitorNode(ctx context.Context, req *pb.FindEnabledMonitorNodeRequest) (*pb.FindEnabledMonitorNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedMonitorNodeDAO.FindEnabledMonitorNode(tx, req.NodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledMonitorNodeResponse{Node: nil}, nil
	}

	result := &pb.MonitorNode{
		Id:          int64(node.Id),
		IsOn:        node.IsOn == 1,
		UniqueId:    node.UniqueId,
		Secret:      node.Secret,
		Name:        node.Name,
		Description: node.Description,
	}
	return &pb.FindEnabledMonitorNodeResponse{Node: result}, nil
}

// 获取当前监控节点的版本
func (this *MonitorNodeService) FindCurrentMonitorNode(ctx context.Context, req *pb.FindCurrentMonitorNodeRequest) (*pb.FindCurrentMonitorNodeResponse, error) {
	_, err := this.ValidateMonitor(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("context: need 'nodeId'")
	}
	nodeIds := md.Get("nodeid")
	if len(nodeIds) == 0 {
		return nil, errors.New("invalid 'nodeId'")
	}
	nodeId := nodeIds[0]
	node, err := models.SharedMonitorNodeDAO.FindEnabledMonitorNodeWithUniqueId(tx, nodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindCurrentMonitorNodeResponse{Node: nil}, nil
	}

	result := &pb.MonitorNode{
		Id:          int64(node.Id),
		IsOn:        node.IsOn == 1,
		UniqueId:    node.UniqueId,
		Secret:      node.Secret,
		Name:        node.Name,
		Description: node.Description,
	}
	return &pb.FindCurrentMonitorNodeResponse{Node: result}, nil
}

// 更新节点状态
func (this *MonitorNodeService) UpdateMonitorNodeStatus(ctx context.Context, req *pb.UpdateMonitorNodeStatusRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeMonitor)
	if err != nil {
		return nil, err
	}

	if req.NodeId > 0 {
		nodeId = req.NodeId
	}

	if nodeId <= 0 {
		return nil, errors.New("'nodeId' should be greater than 0")
	}

	tx := this.NullTx()

	err = models.SharedMonitorNodeDAO.UpdateNodeStatus(tx, nodeId, req.StatusJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
