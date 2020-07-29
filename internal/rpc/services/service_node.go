package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
)

type NodeService struct {
}

func (this *NodeService) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodeId, err := models.SharedNodeDAO.CreateNode(req.Name, int(req.ClusterId))
	if err != nil {
		return nil, err
	}

	return &pb.CreateNodeResponse{
		NodeId: int64(nodeId),
	}, nil
}

func (this *NodeService) CountAllEnabledNodes(ctx context.Context, req *pb.CountAllEnabledNodesRequest) (*pb.CountAllEnabledNodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedNodeDAO.CountAllEnabledNodes()
	if err != nil {
		return nil, err
	}

	return &pb.CountAllEnabledNodesResponse{Count: count}, nil
}

func (this *NodeService) ListEnabledNodes(ctx context.Context, req *pb.ListEnabledNodesRequest) (*pb.ListEnabledNodesResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}
	nodes, err := models.SharedNodeDAO.ListEnabledNodes(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.Node{}
	for _, node := range nodes {
		// 集群信息
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(int64(node.ClusterId))
		if err != nil {
			return nil, err
		}

		result = append(result, &pb.Node{
			Id:   int64(node.Id),
			Name: node.Name,
			Cluster: &pb.NodeCluster{
				Id:   int64(node.ClusterId),
				Name: clusterName,
			},
		})
	}

	return &pb.ListEnabledNodesResponse{
		Nodes: result,
	}, nil
}
