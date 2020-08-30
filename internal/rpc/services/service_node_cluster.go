package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
)

type NodeClusterService struct {
}

// 创建集群
func (this *NodeClusterService) CreateNodeCluster(ctx context.Context, req *pb.CreateNodeClusterRequest) (*pb.CreateNodeClusterResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	clusterId, err := models.SharedNodeClusterDAO.CreateCluster(req.Name, req.GrantId, req.InstallDir)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNodeClusterResponse{ClusterId: clusterId}, nil
}

// 修改集群
func (this *NodeClusterService) UpdateNodeCluster(ctx context.Context, req *pb.UpdateNodeClusterRequest) (*pb.UpdateNodeClusterResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeClusterDAO.UpdateCluster(req.ClusterId, req.Name, req.GrantId, req.InstallDir)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateNodeClusterResponse{}, nil
}

// 禁用集群
func (this *NodeClusterService) DisableNodeCluster(ctx context.Context, req *pb.DisableNodeClusterRequest) (*pb.DisableNodeClusterResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeClusterDAO.DisableNodeCluster(req.ClusterId)
	if err != nil {
		return nil, err
	}

	return &pb.DisableNodeClusterResponse{}, nil
}

// 查找单个集群
func (this *NodeClusterService) FindEnabledNodeCluster(ctx context.Context, req *pb.FindEnabledNodeClusterRequest) (*pb.FindEnabledNodeClusterResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(req.ClusterId)
	if err != nil {
		return nil, err
	}

	if cluster == nil {
		return &pb.FindEnabledNodeClusterResponse{}, nil
	}

	return &pb.FindEnabledNodeClusterResponse{Cluster: &pb.NodeCluster{
		Id:         int64(cluster.Id),
		Name:       cluster.Name,
		CreatedAt:  int64(cluster.CreatedAt),
		InstallDir: cluster.InstallDir,
		GrantId:    int64(cluster.GrantId),
	}}, nil
}

// 查找所有可用的集群
func (this *NodeClusterService) FindAllEnabledNodeClusters(ctx context.Context, req *pb.FindAllEnabledNodeClustersRequest) (*pb.FindAllEnabledNodeClustersResponse, error) {
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
func (this *NodeClusterService) FindAllChangedNodeClusters(ctx context.Context, req *pb.FindAllChangedNodeClustersRequest) (*pb.FindAllChangedNodeClustersResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	clusterIds, err := models.SharedNodeDAO.FindChangedClusterIds()
	if err != nil {
		return nil, err
	}
	if len(clusterIds) == 0 {
		return &pb.FindAllChangedNodeClustersResponse{
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
	return &pb.FindAllChangedNodeClustersResponse{Clusters: result}, nil
}

// 计算所有集群数量
func (this *NodeClusterService) CountAllEnabledNodeClusters(ctx context.Context, req *pb.CountAllEnabledNodeClustersRequest) (*pb.CountAllEnabledNodeClustersResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedNodeClusterDAO.CountAllEnabledClusters()
	if err != nil {
		return nil, err
	}

	return &pb.CountAllEnabledNodeClustersResponse{Count: count}, nil
}

// 列出单页集群
func (this *NodeClusterService) ListEnabledNodeClusters(ctx context.Context, req *pb.ListEnabledNodeClustersRequest) (*pb.ListEnabledNodeClustersResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	clusters, err := models.SharedNodeClusterDAO.ListEnabledClusters(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.NodeCluster{}
	for _, cluster := range clusters {
		result = append(result, &pb.NodeCluster{
			Id:         int64(cluster.Id),
			Name:       cluster.Name,
			CreatedAt:  int64(cluster.CreatedAt),
			GrantId:    int64(cluster.GrantId),
			InstallDir: cluster.InstallDir,
		})
	}

	return &pb.ListEnabledNodeClustersResponse{Clusters: result}, nil
}
