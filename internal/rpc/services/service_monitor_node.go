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

// CreateMonitorNode 创建监控节点
func (this *MonitorNodeService) CreateMonitorNode(ctx context.Context, req *pb.CreateMonitorNodeRequest) (*pb.CreateMonitorNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodeId, err := models.SharedMonitorNodeDAO.CreateMonitorNode(tx, req.Name, req.Description, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMonitorNodeResponse{MonitorNodeId: nodeId}, nil
}

// UpdateMonitorNode 修改监控节点
func (this *MonitorNodeService) UpdateMonitorNode(ctx context.Context, req *pb.UpdateMonitorNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedMonitorNodeDAO.UpdateMonitorNode(tx, req.MonitorNodeId, req.Name, req.Description, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteMonitorNode 删除监控节点
func (this *MonitorNodeService) DeleteMonitorNode(ctx context.Context, req *pb.DeleteMonitorNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedMonitorNodeDAO.DisableMonitorNode(tx, req.MonitorNodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledMonitorNodes 列出所有可用监控节点
func (this *MonitorNodeService) FindAllEnabledMonitorNodes(ctx context.Context, req *pb.FindAllEnabledMonitorNodesRequest) (*pb.FindAllEnabledMonitorNodesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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
			IsOn:        node.IsOn,
			UniqueId:    node.UniqueId,
			Secret:      node.Secret,
			Name:        node.Name,
			Description: node.Description,
		})
	}

	return &pb.FindAllEnabledMonitorNodesResponse{MonitorNodes: result}, nil
}

// CountAllEnabledMonitorNodes 计算监控节点数量
func (this *MonitorNodeService) CountAllEnabledMonitorNodes(ctx context.Context, req *pb.CountAllEnabledMonitorNodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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

// ListEnabledMonitorNodes 列出单页的监控节点
func (this *MonitorNodeService) ListEnabledMonitorNodes(ctx context.Context, req *pb.ListEnabledMonitorNodesRequest) (*pb.ListEnabledMonitorNodesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
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
			IsOn:        node.IsOn,
			UniqueId:    node.UniqueId,
			Secret:      node.Secret,
			Name:        node.Name,
			Description: node.Description,
			StatusJSON:  node.Status,
		})
	}

	return &pb.ListEnabledMonitorNodesResponse{MonitorNodes: result}, nil
}

// FindEnabledMonitorNode 根据ID查找节点
func (this *MonitorNodeService) FindEnabledMonitorNode(ctx context.Context, req *pb.FindEnabledMonitorNodeRequest) (*pb.FindEnabledMonitorNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedMonitorNodeDAO.FindEnabledMonitorNode(tx, req.MonitorNodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledMonitorNodeResponse{MonitorNode: nil}, nil
	}

	result := &pb.MonitorNode{
		Id:          int64(node.Id),
		IsOn:        node.IsOn,
		UniqueId:    node.UniqueId,
		Secret:      node.Secret,
		Name:        node.Name,
		Description: node.Description,
	}
	return &pb.FindEnabledMonitorNodeResponse{MonitorNode: result}, nil
}

// FindCurrentMonitorNode 获取当前监控节点的版本
func (this *MonitorNodeService) FindCurrentMonitorNode(ctx context.Context, req *pb.FindCurrentMonitorNodeRequest) (*pb.FindCurrentMonitorNodeResponse, error) {
	_, err := this.ValidateMonitorNode(ctx)
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
		return &pb.FindCurrentMonitorNodeResponse{MonitorNode: nil}, nil
	}

	result := &pb.MonitorNode{
		Id:          int64(node.Id),
		IsOn:        node.IsOn,
		UniqueId:    node.UniqueId,
		Secret:      node.Secret,
		Name:        node.Name,
		Description: node.Description,
	}
	return &pb.FindCurrentMonitorNodeResponse{MonitorNode: result}, nil
}

// UpdateMonitorNodeStatus 更新节点状态
func (this *MonitorNodeService) UpdateMonitorNodeStatus(ctx context.Context, req *pb.UpdateMonitorNodeStatusRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeMonitor)
	if err != nil {
		return nil, err
	}

	if req.MonitorNodeId > 0 {
		nodeId = req.MonitorNodeId
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
