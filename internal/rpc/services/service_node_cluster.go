package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
)

type NodeClusterService struct {
}

func (this *NodeClusterService) FindAllEnabledClusters(ctx context.Context, req *pb.FindAllEnabledNodeClustersRequest) (*pb.FindAllEnabledNodeClustersResponse, error) {
	_ = req

	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	clusters, err := models.SharedNodeClusterDAO.FindAllEnableClusters()
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:        int64(cluster.Id),
			Name:      cluster.Name,
			CreatedAt: int64(cluster.CreatedAt),
		})
	}

	return &pb.FindAllEnabledNodeClustersResponse{
		Clusters: result,
	}, nil
}

// 查找所有变更的集群
func (this *NodeClusterService) FindAllChangedClusters(ctx context.Context, req *pb.FindAllChangedClustersRequest) (*pb.FindAllChangedClustersResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	clusterIds, err := models.SharedNodeDAO.FindChangedClusterIds()
	if err != nil {
		return nil, err
	}
	if len(clusterIds) == 0 {
		return &pb.FindAllChangedClustersResponse{
			Clusters: []*pb.NodeCluster{},
		}, nil
	}
	result := []*pb.NodeCluster{}
	for _, clusterId := range clusterIds {
		cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(clusterId)
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			continue
		}
		result = append(result, &pb.NodeCluster{
			Id:        int64(cluster.Id),
			Name:      cluster.Name,
			CreatedAt: int64(cluster.CreatedAt),
		})
	}
	return &pb.FindAllChangedClustersResponse{Clusters: result}, nil
}
