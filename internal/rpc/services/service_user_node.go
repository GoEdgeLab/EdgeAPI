package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"google.golang.org/grpc/metadata"
	"time"
)

type UserNodeService struct {
	BaseService
}

// CreateUserNode 创建用户节点
func (this *UserNodeService) CreateUserNode(ctx context.Context, req *pb.CreateUserNodeRequest) (*pb.CreateUserNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodeId, err := models.SharedUserNodeDAO.CreateUserNode(tx, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserNodeResponse{UserNodeId: nodeId}, nil
}

// UpdateUserNode 修改用户节点
func (this *UserNodeService) UpdateUserNode(ctx context.Context, req *pb.UpdateUserNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedUserNodeDAO.UpdateUserNode(tx, req.UserNodeId, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteUserNode 删除用户节点
func (this *UserNodeService) DeleteUserNode(ctx context.Context, req *pb.DeleteUserNodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedUserNodeDAO.DisableUserNode(tx, req.UserNodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledUserNodes 列出所有可用用户节点
func (this *UserNodeService) FindAllEnabledUserNodes(ctx context.Context, req *pb.FindAllEnabledUserNodesRequest) (*pb.FindAllEnabledUserNodesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedUserNodeDAO.FindAllEnabledUserNodes(tx)
	if err != nil {
		return nil, err
	}

	result := []*pb.UserNode{}
	for _, node := range nodes {
		accessAddrs, err := node.DecodeAccessAddrStrings()
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.UserNode{
			Id:              int64(node.Id),
			IsOn:            node.IsOn,
			UniqueId:        node.UniqueId,
			Secret:          node.Secret,
			Name:            node.Name,
			Description:     node.Description,
			HttpJSON:        node.Http,
			HttpsJSON:       node.Https,
			AccessAddrsJSON: node.AccessAddrs,
			AccessAddrs:     accessAddrs,
		})
	}

	return &pb.FindAllEnabledUserNodesResponse{UserNodes: result}, nil
}

// CountAllEnabledUserNodes 计算用户节点数量
func (this *UserNodeService) CountAllEnabledUserNodes(ctx context.Context, req *pb.CountAllEnabledUserNodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedUserNodeDAO.CountAllEnabledUserNodes(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledUserNodes 列出单页的用户节点
func (this *UserNodeService) ListEnabledUserNodes(ctx context.Context, req *pb.ListEnabledUserNodesRequest) (*pb.ListEnabledUserNodesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedUserNodeDAO.ListEnabledUserNodes(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.UserNode{}
	for _, node := range nodes {
		accessAddrs, err := node.DecodeAccessAddrStrings()
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.UserNode{
			Id:              int64(node.Id),
			IsOn:            node.IsOn,
			UniqueId:        node.UniqueId,
			Secret:          node.Secret,
			Name:            node.Name,
			Description:     node.Description,
			HttpJSON:        node.Http,
			HttpsJSON:       node.Https,
			AccessAddrsJSON: node.AccessAddrs,
			AccessAddrs:     accessAddrs,
			StatusJSON:      node.Status,
		})
	}

	return &pb.ListEnabledUserNodesResponse{UserNodes: result}, nil
}

// FindEnabledUserNode 根据ID查找节点
func (this *UserNodeService) FindEnabledUserNode(ctx context.Context, req *pb.FindEnabledUserNodeRequest) (*pb.FindEnabledUserNodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedUserNodeDAO.FindEnabledUserNode(tx, req.UserNodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledUserNodeResponse{UserNode: nil}, nil
	}

	accessAddrs, err := node.DecodeAccessAddrStrings()
	if err != nil {
		return nil, err
	}

	result := &pb.UserNode{
		Id:              int64(node.Id),
		IsOn:            node.IsOn,
		UniqueId:        node.UniqueId,
		Secret:          node.Secret,
		Name:            node.Name,
		Description:     node.Description,
		HttpJSON:        node.Http,
		HttpsJSON:       node.Https,
		AccessAddrsJSON: node.AccessAddrs,
		AccessAddrs:     accessAddrs,
	}
	return &pb.FindEnabledUserNodeResponse{UserNode: result}, nil
}

// FindCurrentUserNode 获取当前用户节点的版本
func (this *UserNodeService) FindCurrentUserNode(ctx context.Context, req *pb.FindCurrentUserNodeRequest) (*pb.FindCurrentUserNodeResponse, error) {
	_, err := this.ValidateUserNode(ctx)
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
	node, err := models.SharedUserNodeDAO.FindEnabledUserNodeWithUniqueId(tx, nodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindCurrentUserNodeResponse{UserNode: nil}, nil
	}

	accessAddrs, err := node.DecodeAccessAddrStrings()
	if err != nil {
		return nil, err
	}

	result := &pb.UserNode{
		Id:              int64(node.Id),
		IsOn:            node.IsOn,
		UniqueId:        node.UniqueId,
		Secret:          node.Secret,
		Name:            node.Name,
		Description:     node.Description,
		HttpJSON:        node.Http,
		HttpsJSON:       node.Https,
		AccessAddrsJSON: node.AccessAddrs,
		AccessAddrs:     accessAddrs,
	}
	return &pb.FindCurrentUserNodeResponse{UserNode: result}, nil
}

// UpdateUserNodeStatus 更新节点状态
func (this *UserNodeService) UpdateUserNodeStatus(ctx context.Context, req *pb.UpdateUserNodeStatusRequest) (*pb.RPCSuccess, error) {
	// 校验节点
	_, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	if req.UserNodeId > 0 {
		nodeId = req.UserNodeId
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
	err = models.SharedUserNodeDAO.UpdateNodeStatus(tx, nodeId, nodeStatus)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledUserNodesWithSSLCertId 计算使用某个SSL证书的用户节点数量
func (this *UserNodeService) CountAllEnabledUserNodesWithSSLCertId(ctx context.Context, req *pb.CountAllEnabledUserNodesWithSSLCertIdRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	policyIds, err := models.SharedSSLPolicyDAO.FindAllEnabledPolicyIdsWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}
	if len(policyIds) == 0 {
		return this.SuccessCount(0)
	}

	count, err := models.SharedUserNodeDAO.CountAllEnabledUserNodesWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}
