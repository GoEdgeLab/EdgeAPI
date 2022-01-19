package services

import (
	"context"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
)

type APINodeService struct {
	BaseService
}

// CreateAPINode 创建API节点
func (this *APINodeService) CreateAPINode(ctx context.Context, req *pb.CreateAPINodeRequest) (*pb.CreateAPINodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodeId, err := models.SharedAPINodeDAO.CreateAPINode(tx, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.RestIsOn, req.RestHTTPJSON, req.RestHTTPSJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return &pb.CreateAPINodeResponse{ApiNodeId: nodeId}, nil
}

// UpdateAPINode 修改API节点
func (this *APINodeService) UpdateAPINode(ctx context.Context, req *pb.UpdateAPINodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedAPINodeDAO.UpdateAPINode(tx, req.ApiNodeId, req.Name, req.Description, req.HttpJSON, req.HttpsJSON, req.RestIsOn, req.RestHTTPJSON, req.RestHTTPSJSON, req.AccessAddrsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteAPINode 删除API节点
func (this *APINodeService) DeleteAPINode(ctx context.Context, req *pb.DeleteAPINodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedAPINodeDAO.DisableAPINode(tx, req.ApiNodeId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledAPINodes 列出所有可用API节点
func (this *APINodeService) FindAllEnabledAPINodes(ctx context.Context, req *pb.FindAllEnabledAPINodesRequest) (*pb.FindAllEnabledAPINodesResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser, rpcutils.UserTypeNode, rpcutils.UserTypeMonitor, rpcutils.UserTypeDNS, rpcutils.UserTypeAuthority)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedAPINodeDAO.FindAllEnabledAPINodes(tx)
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

	return &pb.FindAllEnabledAPINodesResponse{ApiNodes: result}, nil
}

// CountAllEnabledAPINodes 计算API节点数量
func (this *APINodeService) CountAllEnabledAPINodes(ctx context.Context, req *pb.CountAllEnabledAPINodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedAPINodeDAO.CountAllEnabledAPINodes(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// CountAllEnabledAndOnAPINodes 计算API节点数量
func (this *APINodeService) CountAllEnabledAndOnAPINodes(ctx context.Context, req *pb.CountAllEnabledAndOnAPINodesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedAPINodeDAO.CountAllEnabledAndOnAPINodes(tx)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledAPINodes 列出单页的API节点
func (this *APINodeService) ListEnabledAPINodes(ctx context.Context, req *pb.ListEnabledAPINodesRequest) (*pb.ListEnabledAPINodesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	nodes, err := models.SharedAPINodeDAO.ListEnabledAPINodes(tx, req.Offset, req.Size)
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

	return &pb.ListEnabledAPINodesResponse{ApiNodes: result}, nil
}

// FindEnabledAPINode 根据ID查找节点
func (this *APINodeService) FindEnabledAPINode(ctx context.Context, req *pb.FindEnabledAPINodeRequest) (*pb.FindEnabledAPINodeResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	node, err := models.SharedAPINodeDAO.FindEnabledAPINode(tx, req.ApiNodeId, nil)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return &pb.FindEnabledAPINodeResponse{ApiNode: nil}, nil
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
	return &pb.FindEnabledAPINodeResponse{ApiNode: result}, nil
}

// FindCurrentAPINodeVersion 获取当前API节点的版本
func (this *APINodeService) FindCurrentAPINodeVersion(ctx context.Context, req *pb.FindCurrentAPINodeVersionRequest) (*pb.FindCurrentAPINodeVersionResponse, error) {
	_, _, _, err := rpcutils.ValidateRequest(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.FindCurrentAPINodeVersionResponse{Version: teaconst.Version}, nil
}

// FindCurrentAPINode 获取当前API节点的信息
func (this *APINodeService) FindCurrentAPINode(ctx context.Context, req *pb.FindCurrentAPINodeRequest) (*pb.FindCurrentAPINodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var nodeId = teaconst.NodeId
	var tx *dbs.Tx
	node, err := models.SharedAPINodeDAO.FindEnabledAPINode(tx, nodeId, nil)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return &pb.FindCurrentAPINodeResponse{ApiNode: nil}, nil
	}

	accessAddrs, err := node.DecodeAccessAddrStrings()
	if err != nil {
		return nil, err
	}

	return &pb.FindCurrentAPINodeResponse{ApiNode: &pb.APINode{
		Id:              int64(node.Id),
		IsOn:            node.IsOn == 1,
		NodeClusterId:   0,
		UniqueId:        "",
		Secret:          "",
		Name:            "",
		Description:     "",
		HttpJSON:        nil,
		HttpsJSON:       nil,
		RestIsOn:        false,
		RestHTTPJSON:    nil,
		RestHTTPSJSON:   nil,
		AccessAddrsJSON: []byte(node.AccessAddrs),
		AccessAddrs:     accessAddrs,
		StatusJSON:      nil,
	}}, nil
}

// CountAllEnabledAPINodesWithSSLCertId 计算使用某个SSL证书的API节点数量
func (this *APINodeService) CountAllEnabledAPINodesWithSSLCertId(ctx context.Context, req *pb.CountAllEnabledAPINodesWithSSLCertIdRequest) (*pb.RPCCountResponse, error) {
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

	count, err := models.SharedAPINodeDAO.CountAllEnabledAPINodesWithSSLPolicyIds(tx, policyIds)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// DebugAPINode 修改调试模式状态
func (this *APINodeService) DebugAPINode(ctx context.Context, req *pb.DebugAPINodeRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	teaconst.Debug = req.Debug
	return this.Success()
}
