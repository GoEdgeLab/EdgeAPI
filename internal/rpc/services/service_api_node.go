package services

import (
	"context"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type APINodeService struct {
	BaseService
}

// 创建API节点
func (this *APINodeService) CreateAPINode(ctx context.Context, req *pb.CreateAPINodeRequest) (*pb.CreateAPINodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodeId, err := models.SharedAPINodeDAO.CreateAPINode(req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.RestIsOn, req.RestHTTPJSON, req.RestHTTPSJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateAPINodeResponse{NodeId: nodeId}, nil
}

// 修改API节点
func (this *APINodeService) UpdateAPINode(ctx context.Context, req *pb.UpdateAPINodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedAPINodeDAO.UpdateAPINode(req.NodeId, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.RestIsOn, req.RestHTTPJSON, req.RestHTTPSJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 删除API节点
func (this *APINodeService) DeleteAPINode(ctx context.Context, req *pb.DeleteAPINodeRequest) (*pb.RPCSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedAPINodeDAO.DisableAPINode(req.NodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 列出所有可用API节点
func (this *APINodeService) FindAllEnabledAPINodes(ctx context.Context, req *pb.FindAllEnabledAPINodesRequest) (*pb.FindAllEnabledAPINodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodes, err := models.SharedAPINodeDAO.FindAllEnabledAPINodes()
	if err != nil {
		return nil, err
	}

	result := []*pb.APINode{}
	for _, node := range nodes {
		accessAddrs, err := node.DecodeAccessAddrStrings()
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.APINode{
			Id:              int64(node.Id),
			IsOn:            node.IsOn == 1,
			NodeClusterId:   int64(node.ClusterId),
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

	return &pb.FindAllEnabledAPINodesResponse{Nodes: result}, nil
}

// 计算API节点数量
func (this *APINodeService) CountAllEnabledAPINodes(ctx context.Context, req *pb.CountAllEnabledAPINodesRequest) (*pb.RPCCountResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedAPINodeDAO.CountAllEnabledAPINodes()
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// 列出单页的API节点
func (this *APINodeService) ListEnabledAPINodes(ctx context.Context, req *pb.ListEnabledAPINodesRequest) (*pb.ListEnabledAPINodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodes, err := models.SharedAPINodeDAO.ListEnabledAPINodes(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.APINode{}
	for _, node := range nodes {
		accessAddrs, err := node.DecodeAccessAddrStrings()
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.APINode{
			Id:              int64(node.Id),
			IsOn:            node.IsOn == 1,
			NodeClusterId:   int64(node.ClusterId),
			UniqueId:        node.UniqueId,
			Secret:          node.Secret,
			Name:            node.Name,
			Description:     node.Description,
			HttpJSON:        []byte(node.Http),
			HttpsJSON:       []byte(node.Https),
			RestIsOn:        node.RestIsOn == 1,
			RestHTTPJSON:    []byte(node.RestHTTP),
			RestHTTPSJSON:   []byte(node.RestHTTPS),
			AccessAddrsJSON: []byte(node.AccessAddrs),
			AccessAddrs:     accessAddrs,
			StatusJSON:      []byte(node.Status),
		})
	}

	return &pb.ListEnabledAPINodesResponse{Nodes: result}, nil
}

// 根据ID查找节点
func (this *APINodeService) FindEnabledAPINode(ctx context.Context, req *pb.FindEnabledAPINodeRequest) (*pb.FindEnabledAPINodeResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	node, err := models.SharedAPINodeDAO.FindEnabledAPINode(req.NodeId)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledAPINodeResponse{Node: nil}, nil
	}

	accessAddrs, err := node.DecodeAccessAddrStrings()
	if err != nil {
		return nil, err
	}

	result := &pb.APINode{
		Id:              int64(node.Id),
		IsOn:            node.IsOn == 1,
		NodeClusterId:   int64(node.ClusterId),
		UniqueId:        node.UniqueId,
		Secret:          node.Secret,
		Name:            node.Name,
		Description:     node.Description,
		HttpJSON:        []byte(node.Http),
		HttpsJSON:       []byte(node.Https),
		RestIsOn:        node.RestIsOn == 1,
		RestHTTPJSON:    []byte(node.RestHTTP),
		RestHTTPSJSON:   []byte(node.RestHTTPS),
		AccessAddrsJSON: []byte(node.AccessAddrs),
		AccessAddrs:     accessAddrs,
	}
	return &pb.FindEnabledAPINodeResponse{Node: result}, nil
}

// 获取当前API节点的版本
func (this *APINodeService) FindCurrentAPINodeVersion(ctx context.Context, req *pb.FindCurrentAPINodeVersionRequest) (*pb.FindCurrentAPINodeVersionResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.FindCurrentAPINodeVersionResponse{Version: teaconst.Version}, nil
}
