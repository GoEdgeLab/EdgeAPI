// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// NodeClusterMetricItemService 集群指标
type NodeClusterMetricItemService struct {
	BaseService
}

// EnableNodeClusterMetricItem 启用某个指标
func (this *NodeClusterMetricItemService) EnableNodeClusterMetricItem(ctx context.Context, req *pb.EnableNodeClusterMetricItemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	exists, err := models.SharedNodeClusterMetricItemDAO.ExistsClusterItem(tx, req.NodeClusterId, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = models.SharedNodeClusterMetricItemDAO.EnableClusterItem(tx, req.NodeClusterId, req.MetricItemId)
		if err != nil {
			return nil, err
		}
	}
	return this.Success()
}

// DisableNodeClusterMetricItem 禁用某个指标
func (this *NodeClusterMetricItemService) DisableNodeClusterMetricItem(ctx context.Context, req *pb.DisableNodeClusterMetricItemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeClusterMetricItemDAO.DisableClusterItem(tx, req.NodeClusterId, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllNodeClusterMetricItems 查找集群中所有指标
func (this *NodeClusterMetricItemService) FindAllNodeClusterMetricItems(ctx context.Context, req *pb.FindAllNodeClusterMetricItemsRequest) (*pb.FindAllNodeClusterMetricItemsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	clusterItems, err := models.SharedNodeClusterMetricItemDAO.FindAllClusterItems(tx, req.NodeClusterId, req.Category)
	if err != nil {
		return nil, err
	}
	var pbItems = []*pb.MetricItem{}
	for _, clusterItem := range clusterItems {
		item, err := models.SharedMetricItemDAO.FindEnabledMetricItem(tx, int64(clusterItem.ItemId))
		if err != nil {
			return nil, err
		}
		if item != nil {
			pbItems = append(pbItems, &pb.MetricItem{
				Id:         int64(item.Id),
				IsOn:       item.IsOn,
				Code:       item.Code,
				Category:   item.Category,
				Name:       item.Name,
				Keys:       item.DecodeKeys(),
				Period:     types.Int32(item.Period),
				PeriodUnit: item.PeriodUnit,
				Value:      item.Value,
				IsPublic:   item.IsPublic,
			})
		}
	}
	return &pb.FindAllNodeClusterMetricItemsResponse{MetricItems: pbItems}, nil
}

// ExistsNodeClusterMetricItem 检查是否已添加某个指标
func (this *NodeClusterMetricItemService) ExistsNodeClusterMetricItem(ctx context.Context, req *pb.ExistsNodeClusterMetricItemRequest) (*pb.RPCExists, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	b, err := models.SharedNodeClusterMetricItemDAO.ExistsClusterItem(tx, req.NodeClusterId, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	return this.Exists(b)
}

// FindAllNodeClustersWithMetricItemId 查找使用指标的集群
func (this *NodeClusterMetricItemService) FindAllNodeClustersWithMetricItemId(ctx context.Context, req *pb.FindAllNodeClustersWithMetricItemIdRequest) (*pb.FindAllNodeClustersWithMetricItemIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	clusterIds, err := models.SharedNodeClusterMetricItemDAO.FindAllClusterIdsWithItemId(tx, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	var pbClusters []*pb.NodeCluster
	for _, clusterId := range clusterIds {
		cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, clusterId)
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			continue
		}
		pbClusters = append(pbClusters, &pb.NodeCluster{
			Id:   int64(cluster.Id),
			Name: cluster.Name,
		})
	}

	return &pb.FindAllNodeClustersWithMetricItemIdResponse{NodeClusters: pbClusters}, nil
}
