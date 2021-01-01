package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"google.golang.org/grpc/metadata"
)

type UserNodeService struct {
	BaseService
}

// 创建用户节点
func (this *UserNodeService) CreateUserNode(ctx context.Context, req *pb.CreateUserNodeRequest) (*pb.CreateUserNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodeId, err := models.SharedUserNodeDAO.CreateUserNode(tx, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserNodeResponse{NodeId: nodeId}, nil
}

// 修改用户节点
func (this *UserNodeService) UpdateUserNode(ctx context.Context, req *pb.UpdateUserNodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedUserNodeDAO.UpdateUserNode(tx, req.NodeId, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 删除用户节点
func (this *UserNodeService) DeleteUserNode(ctx context.Context, req *pb.DeleteUserNodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedUserNodeDAO.DisableUserNode(tx, req.NodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 列出所有可用用户节点
func (this *UserNodeService) FindAllEnabledUserNodes(ctx context.Context, req *pb.FindAllEnabledUserNodesRequest) (*pb.FindAllEnabledUserNodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
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
			IsOn:            node.IsOn == 1,
			UniqueId:        node.UniqueId,
			Secret:          node.Secret,
			Name:            node.Name,
			Description:     node.Description,
			HttpJSON:        []byte(node.Http),
			HttpsJSON:       []byte(node.Https),
			AccessAddrsJSON: []byte(node.AccessAddrs),
			AccessAddrs:     accessAddrs,
		})
	}

	return &pb.FindAllEnabledUserNodesResponse{Nodes: result}, nil
}

// 计算用户节点数量
func (this *UserNodeService) CountAllEnabledUserNodes(ctx context.Context, req *pb.CountAllEnabledUserNodesRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
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

// 列出单页的用户节点
func (this *UserNodeService) ListEnabledUserNodes(ctx context.Context, req *pb.ListEnabledUserNodesRequest) (*pb.ListEnabledUserNodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
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
			IsOn:            node.IsOn == 1,
			UniqueId:        node.UniqueId,
			Secret:          node.Secret,
			Name:            node.Name,
			Description:     node.Description,
			HttpJSON:        []byte(node.Http),
			HttpsJSON:       []byte(node.Https),
			AccessAddrsJSON: []byte(node.AccessAddrs),
			AccessAddrs:     accessAddrs,
		})
	}

	return &pb.ListEnabledUserNodesResponse{Nodes: result}, nil
}

// 根据ID查找节点
func (this *UserNodeService) FindEnabledUserNode(ctx context.Context, req *pb.FindEnabledUserNodeRequest) (*pb.FindEnabledUserNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedUserNodeDAO.FindEnabledUserNode(tx, req.NodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledUserNodeResponse{Node: nil}, nil
	}

	accessAddrs, err := node.DecodeAccessAddrStrings()
	if err != nil {
		return nil, err
	}

	result := &pb.UserNode{
		Id:              int64(node.Id),
		IsOn:            node.IsOn == 1,
		UniqueId:        node.UniqueId,
		Secret:          node.Secret,
		Name:            node.Name,
		Description:     node.Description,
		HttpJSON:        []byte(node.Http),
		HttpsJSON:       []byte(node.Https),
		AccessAddrsJSON: []byte(node.AccessAddrs),
		AccessAddrs:     accessAddrs,
	}
	return &pb.FindEnabledUserNodeResponse{Node: result}, nil
}

// 获取当前用户节点的版本
func (this *UserNodeService) FindCurrentUserNode(ctx context.Context, req *pb.FindCurrentUserNodeRequest) (*pb.FindCurrentUserNodeResponse, error) {
	_, err := this.ValidateUser(ctx)
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
		return &pb.FindCurrentUserNodeResponse{Node: nil}, nil
	}

	accessAddrs, err := node.DecodeAccessAddrStrings()
	if err != nil {
		return nil, err
	}

	result := &pb.UserNode{
		Id:              int64(node.Id),
		IsOn:            node.IsOn == 1,
		UniqueId:        node.UniqueId,
		Secret:          node.Secret,
		Name:            node.Name,
		Description:     node.Description,
		HttpJSON:        []byte(node.Http),
		HttpsJSON:       []byte(node.Https),
		AccessAddrsJSON: []byte(node.AccessAddrs),
		AccessAddrs:     accessAddrs,
	}
	return &pb.FindCurrentUserNodeResponse{Node: result}, nil
}
