package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/authority"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"google.golang.org/grpc/metadata"
	"time"
)

type AuthorityNodeService struct {
	BaseService
}

// CreateAuthorityNode 创建认证节点
func (this *AuthorityNodeService) CreateAuthorityNode(ctx context.Context, req *pb.CreateAuthorityNodeRequest) (*pb.CreateAuthorityNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodeId, err := authority.SharedAuthorityNodeDAO.CreateAuthorityNode(tx, req.Name, req.Description, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateAuthorityNodeResponse{AuthorityNodeId: nodeId}, nil
}

// UpdateAuthorityNode 修改认证节点
func (this *AuthorityNodeService) UpdateAuthorityNode(ctx context.Context, req *pb.UpdateAuthorityNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = authority.SharedAuthorityNodeDAO.UpdateAuthorityNode(tx, req.AuthorityNodeId, req.Name, req.Description, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteAuthorityNode 删除认证节点
func (this *AuthorityNodeService) DeleteAuthorityNode(ctx context.Context, req *pb.DeleteAuthorityNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = authority.SharedAuthorityNodeDAO.DisableAuthorityNode(tx, req.AuthorityNodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledAuthorityNodes 列出所有可用认证节点
func (this *AuthorityNodeService) FindAllEnabledAuthorityNodes(ctx context.Context, req *pb.FindAllEnabledAuthorityNodesRequest) (*pb.FindAllEnabledAuthorityNodesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodes, err := authority.SharedAuthorityNodeDAO.FindAllEnabledAuthorityNodes(tx)
	if err != nil {
		return nil, err
	}

	result := []*pb.AuthorityNode{}
	for _, node := range nodes {
		result = append(result, &pb.AuthorityNode{
			Id:          int64(node.Id),
			IsOn:        node.IsOn,
			UniqueId:    node.UniqueId,
			Secret:      node.Secret,
			Name:        node.Name,
			Description: node.Description,
		})
	}

	return &pb.FindAllEnabledAuthorityNodesResponse{AuthorityNodes: result}, nil
}

// CountAllEnabledAuthorityNodes 计算认证节点数量
func (this *AuthorityNodeService) CountAllEnabledAuthorityNodes(ctx context.Context, req *pb.CountAllEnabledAuthorityNodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := authority.SharedAuthorityNodeDAO.CountAllEnabledAuthorityNodes(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledAuthorityNodes 列出单页的认证节点
func (this *AuthorityNodeService) ListEnabledAuthorityNodes(ctx context.Context, req *pb.ListEnabledAuthorityNodesRequest) (*pb.ListEnabledAuthorityNodesResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	nodes, err := authority.SharedAuthorityNodeDAO.ListEnabledAuthorityNodes(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.AuthorityNode{}
	for _, node := range nodes {
		result = append(result, &pb.AuthorityNode{
			Id:          int64(node.Id),
			IsOn:        node.IsOn,
			UniqueId:    node.UniqueId,
			Secret:      node.Secret,
			Name:        node.Name,
			Description: node.Description,
			StatusJSON:  node.Status,
		})
	}

	return &pb.ListEnabledAuthorityNodesResponse{AuthorityNodes: result}, nil
}

// FindEnabledAuthorityNode 根据ID查找节点
func (this *AuthorityNodeService) FindEnabledAuthorityNode(ctx context.Context, req *pb.FindEnabledAuthorityNodeRequest) (*pb.FindEnabledAuthorityNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	node, err := authority.SharedAuthorityNodeDAO.FindEnabledAuthorityNode(tx, req.AuthorityNodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledAuthorityNodeResponse{AuthorityNode: nil}, nil
	}

	result := &pb.AuthorityNode{
		Id:          int64(node.Id),
		IsOn:        node.IsOn,
		UniqueId:    node.UniqueId,
		Secret:      node.Secret,
		Name:        node.Name,
		Description: node.Description,
	}
	return &pb.FindEnabledAuthorityNodeResponse{AuthorityNode: result}, nil
}

// FindCurrentAuthorityNode 获取当前认证节点的版本
func (this *AuthorityNodeService) FindCurrentAuthorityNode(ctx context.Context, req *pb.FindCurrentAuthorityNodeRequest) (*pb.FindCurrentAuthorityNodeResponse, error) {
	_, err := this.ValidateAuthorityNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("context: need 'nodeId'")
	}
	nodeIds := md.Get("nodeid")
	if len(nodeIds) == 0 {
		return nil, errors.New("invalid 'nodeId'")
	}
	nodeId := nodeIds[0]
	node, err := authority.SharedAuthorityNodeDAO.FindEnabledAuthorityNodeWithUniqueId(tx, nodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindCurrentAuthorityNodeResponse{AuthorityNode: nil}, nil
	}

	result := &pb.AuthorityNode{
		Id:          int64(node.Id),
		IsOn:        node.IsOn,
		UniqueId:    node.UniqueId,
		Secret:      node.Secret,
		Name:        node.Name,
		Description: node.Description,
	}
	return &pb.FindCurrentAuthorityNodeResponse{AuthorityNode: result}, nil
}

// UpdateAuthorityNodeStatus 更新节点状态
func (this *AuthorityNodeService) UpdateAuthorityNodeStatus(ctx context.Context, req *pb.UpdateAuthorityNodeStatusRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeAuthority)
	if err != nil {
		return nil, err
	}

	if req.AuthorityNodeId > 0 {
		nodeId = req.AuthorityNodeId
	}

	if nodeId <= 0 {
		return nil, errors.New("'nodeId' should be greater than 0")
	}

	var tx = this.NullTx()

	// 修改时间戳
	var nodeStatus = &nodeconfigs.NodeStatus{}
	err = json.Unmarshal(req.StatusJSON, nodeStatus)
	if err != nil {
		return nil, errors.New("decode node status json failed: " + err.Error())
	}
	nodeStatus.UpdatedAt = time.Now().Unix()

	// 保存
	err = authority.SharedAuthorityNodeDAO.UpdateNodeStatus(tx, nodeId, nodeStatus)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
