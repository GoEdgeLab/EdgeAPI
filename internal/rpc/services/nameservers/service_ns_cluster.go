// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NSClusterService 域名服务集群相关服务
type NSClusterService struct {
	services.BaseService
}

// CreateNSCluster 创建集群
func (this *NSClusterService) CreateNSCluster(ctx context.Context, req *pb.CreateNSClusterRequest) (*pb.CreateNSClusterResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	clusterId, err := models.SharedNSClusterDAO.CreateCluster(tx, req.Name, req.AccessLogJSON)
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
	err = models.SharedNSClusterDAO.UpdateCluster(tx, req.NsClusterId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindNSClusterAccessLog 查找集群访问日志配置
func (this *NSClusterService) FindNSClusterAccessLog(ctx context.Context, req *pb.FindNSClusterAccessLogRequest) (*pb.FindNSClusterAccessLogResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	accessLogJSON, err := models.SharedNSClusterDAO.FindClusterAccessLog(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.FindNSClusterAccessLogResponse{AccessLogJSON: accessLogJSON}, nil
}

// UpdateNSClusterAccessLog 修改集群访问日志配置
func (this *NSClusterService) UpdateNSClusterAccessLog(ctx context.Context, req *pb.UpdateNSClusterAccessLogRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.UpdateClusterAccessLog(tx, req.NsClusterId, req.AccessLogJSON)
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
	err = models.SharedNSClusterDAO.DisableNSCluster(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}

	// 删除任务
	err = models.SharedNodeTaskDAO.DeleteAllClusterTasks(tx, nodeconfigs.NodeRoleDNS, req.NsClusterId)
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
	cluster, err := models.SharedNSClusterDAO.FindEnabledNSCluster(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return &pb.FindEnabledNSClusterResponse{NsCluster: nil}, nil
	}
	return &pb.FindEnabledNSClusterResponse{NsCluster: &pb.NSCluster{
		Id:         int64(cluster.Id),
		IsOn:       cluster.IsOn,
		Name:       cluster.Name,
		InstallDir: cluster.InstallDir,
	}}, nil
}

// CountAllEnabledNSClusters 计算所有可用集群的数量
func (this *NSClusterService) CountAllEnabledNSClusters(ctx context.Context, req *pb.CountAllEnabledNSClustersRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	count, err := models.SharedNSClusterDAO.CountAllEnabledClusters(tx)
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
	clusters, err := models.SharedNSClusterDAO.ListEnabledClusters(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbClusters = []*pb.NSCluster{}
	for _, cluster := range clusters {
		pbClusters = append(pbClusters, &pb.NSCluster{
			Id:         int64(cluster.Id),
			IsOn:       cluster.IsOn,
			Name:       cluster.Name,
			InstallDir: cluster.InstallDir,
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
	clusters, err := models.SharedNSClusterDAO.FindAllEnabledClusters(tx)
	if err != nil {
		return nil, err
	}
	var pbClusters = []*pb.NSCluster{}
	for _, cluster := range clusters {
		pbClusters = append(pbClusters, &pb.NSCluster{
			Id:         int64(cluster.Id),
			IsOn:       cluster.IsOn,
			Name:       cluster.Name,
			InstallDir: cluster.InstallDir,
		})
	}
	return &pb.FindAllEnabledNSClustersResponse{NsClusters: pbClusters}, nil
}

// UpdateNSClusterRecursionConfig 设置递归DNS配置
func (this *NSClusterService) UpdateNSClusterRecursionConfig(ctx context.Context, req *pb.UpdateNSClusterRecursionConfigRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// 校验配置
	var config = &dnsconfigs.RecursionConfig{}
	err = json.Unmarshal(req.RecursionJSON, config)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNSClusterDAO.UpdateRecursion(tx, req.NsClusterId, req.RecursionJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindNSClusterRecursionConfig 读取递归DNS配置
func (this *NSClusterService) FindNSClusterRecursionConfig(ctx context.Context, req *pb.FindNSClusterRecursionConfigRequest) (*pb.FindNSClusterRecursionConfigResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	recursion, err := models.SharedNSClusterDAO.FindClusterRecursion(tx, req.NsClusterId)
	if err != nil {
		return nil, err
	}
	return &pb.FindNSClusterRecursionConfigResponse{
		RecursionJSON: recursion,
	}, nil
}
