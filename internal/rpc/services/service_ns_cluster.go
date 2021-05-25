// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NSClusterService 域名服务集群相关服务
type NSClusterService struct {
	BaseService
}

// CreateNSCluster 创建集群
func (this *NSClusterService) CreateNSCluster(ctx context.Context, req *pb.CreateNSClusterRequest) (*pb.CreateNSClusterResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	clusterId, err := nameservers.SharedNSClusterDAO.CreateCluster(tx, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSClusterResponse{NsClusterId: clusterId}, nil
}

// UpdateNSCluster 修改集群
func (this *NSClusterService) UpdateNSCluster(ctx context.Context, req *pb.UpdateNSClusterRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = nameservers.SharedNSClusterDAO.UpdateCluster(tx, req.NsClusterId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNSCluster 删除集群
func (this *NSClusterService) DeleteNSCluster(ctx context.Context, req *pb.DeleteNSCluster) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = nameservers.SharedNSClusterDAO.DisableNSCluster(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledNSCluster 查找单个可用集群信息
func (this *NSClusterService) FindEnabledNSCluster(ctx context.Context, req *pb.FindEnabledNSClusterRequest) (*pb.FindEnabledNSClusterResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	cluster, err := nameservers.SharedNSClusterDAO.FindEnabledNSCluster(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return &pb.FindEnabledNSClusterResponse{NsCluster: nil}, nil
	}
	return &pb.FindEnabledNSClusterResponse{NsCluster: &pb.NSCluster{
		Id:   int64(cluster.Id),
		IsOn: cluster.IsOn == 1,
		Name: cluster.Name,
	}}, nil
}

// CountAllEnabledNSClusters 计算所有可用集群的数量
func (this *NSClusterService) CountAllEnabledNSClusters(ctx context.Context, req *pb.CountAllEnabledNSClustersRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	count, err := nameservers.SharedNSClusterDAO.CountAllEnabledClusters(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledNSClusters 列出单页可用集群
func (this *NSClusterService) ListEnabledNSClusters(ctx context.Context, req *pb.ListEnabledNSClustersRequest) (*pb.ListEnabledNSClustersResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	clusters, err := nameservers.SharedNSClusterDAO.ListEnabledNSClusters(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbClusters = []*pb.NSCluster{}
	for _, cluster := range clusters {
		pbClusters = append(pbClusters, &pb.NSCluster{
			Id:   int64(cluster.Id),
			IsOn: cluster.IsOn == 1,
			Name: cluster.Name,
		})
	}
	return &pb.ListEnabledNSClustersResponse{NsClusters: pbClusters}, nil
}

// FindAllEnabledNSClusters 查找所有可用集群
func (this *NSClusterService) FindAllEnabledNSClusters(ctx context.Context, req *pb.FindAllEnabledNSClustersRequest) (*pb.FindAllEnabledNSClustersResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	clusters, err := nameservers.SharedNSClusterDAO.FindAllEnabledNSClusters(tx)
	if err != nil {
		return nil, err
	}
	var pbClusters = []*pb.NSCluster{}
	for _, cluster := range clusters {
		pbClusters = append(pbClusters, &pb.NSCluster{
			Id:   int64(cluster.Id),
			IsOn: cluster.IsOn == 1,
			Name: cluster.Name,
		})
	}
	return &pb.FindAllEnabledNSClustersResponse{NsClusters: pbClusters}, nil
}
