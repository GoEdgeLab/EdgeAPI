// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// NodeThresholdService 节点阈值服务
type NodeThresholdService struct {
	BaseService
}

// CreateNodeThreshold 创建阈值
func (this *NodeThresholdService) CreateNodeThreshold(ctx context.Context, req *pb.CreateNodeThresholdRequest) (*pb.CreateNodeThresholdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	thresholdId, err := models.SharedNodeThresholdDAO.CreateThreshold(tx, rpcutils.UserTypeNode, req.NodeClusterId, req.NodeId, req.Item, req.Param, req.Operator, req.ValueJSON, req.Message, req.SumMethod, req.Duration, req.DurationUnit, req.NotifyDuration)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodeThresholdResponse{NodeThresholdId: thresholdId}, nil
}

// UpdateNodeThreshold 创建阈值
func (this *NodeThresholdService) UpdateNodeThreshold(ctx context.Context, req *pb.UpdateNodeThresholdRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = models.SharedNodeThresholdDAO.UpdateThreshold(tx, req.NodeThresholdId, req.Item, req.Param, req.Operator, req.ValueJSON, req.Message, req.SumMethod, req.Duration, req.DurationUnit, req.NotifyDuration, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNodeThreshold 删除阈值
func (this *NodeThresholdService) DeleteNodeThreshold(ctx context.Context, req *pb.DeleteNodeThresholdRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	err = models.SharedNodeThresholdDAO.DisableNodeThreshold(tx, req.NodeThresholdId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllEnabledNodeThresholds 查询阈值
func (this *NodeThresholdService) FindAllEnabledNodeThresholds(ctx context.Context, req *pb.FindAllEnabledNodeThresholdsRequest) (*pb.FindAllEnabledNodeThresholdsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	pbThresholds := []*pb.NodeThreshold{}
	thresholds, err := models.SharedNodeThresholdDAO.FindAllEnabledThresholds(tx, req.Role, req.NodeClusterId, req.NodeId)
	if err != nil {
		return nil, err
	}
	for _, threshold := range thresholds {
		// 节点信息
		var pbNode *pb.Node = nil
		if threshold.NodeId > 0 {
			nodeName, err := models.SharedNodeDAO.FindNodeName(tx, int64(threshold.NodeId))
			if err != nil {
				return nil, err
			}
			if len(nodeName) == 0 {
				continue
			}
			pbNode = &pb.Node{
				Id:   int64(threshold.NodeId),
				Name: nodeName,
			}
		}

		pbThresholds = append(pbThresholds, &pb.NodeThreshold{
			Id:             int64(threshold.Id),
			ClusterId:      int64(threshold.ClusterId),
			Node:           pbNode,
			Item:           threshold.Item,
			Param:          threshold.Param,
			Operator:       threshold.Operator,
			ValueJSON:      []byte(threshold.Value),
			Message:        threshold.Message,
			Duration:       types.Int32(threshold.Duration),
			DurationUnit:   threshold.DurationUnit,
			SumMethod:      threshold.SumMethod,
			NotifyDuration: int32(threshold.NotifyDuration),
			IsOn:           threshold.IsOn == 1,
		})
	}
	return &pb.FindAllEnabledNodeThresholdsResponse{NodeThresholds: pbThresholds}, nil
}

// CountAllEnabledNodeThresholds 计算阈值数量
func (this *NodeThresholdService) CountAllEnabledNodeThresholds(ctx context.Context, req *pb.CountAllEnabledNodeThresholdsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNodeThresholdDAO.CountAllEnabledThresholds(tx, req.NodeClusterId, req.NodeId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// FindEnabledNodeThreshold 查询单个阈值详情
func (this *NodeThresholdService) FindEnabledNodeThreshold(ctx context.Context, req *pb.FindEnabledNodeThresholdRequest) (*pb.FindEnabledNodeThresholdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	threshold, err := models.SharedNodeThresholdDAO.FindEnabledNodeThreshold(tx, req.NodeThresholdId)
	if err != nil {
		return nil, err
	}
	if threshold == nil {
		return &pb.FindEnabledNodeThresholdResponse{NodeThreshold: nil}, nil
	}

	// 节点信息
	var pbNode *pb.Node = nil
	if threshold.NodeId > 0 {
		nodeName, err := models.SharedNodeDAO.FindNodeName(tx, int64(threshold.NodeId))
		if err != nil {
			return nil, err
		}
		if len(nodeName) == 0 {
			return &pb.FindEnabledNodeThresholdResponse{NodeThreshold: nil}, nil
		}
		pbNode = &pb.Node{
			Id:   int64(threshold.NodeId),
			Name: nodeName,
		}
	}

	return &pb.FindEnabledNodeThresholdResponse{NodeThreshold: &pb.NodeThreshold{
		Id:             int64(threshold.Id),
		ClusterId:      int64(threshold.ClusterId),
		Node:           pbNode,
		Item:           threshold.Item,
		Param:          threshold.Param,
		Operator:       threshold.Operator,
		ValueJSON:      []byte(threshold.Value),
		Message:        threshold.Message,
		Duration:       types.Int32(threshold.Duration),
		DurationUnit:   threshold.DurationUnit,
		SumMethod:      threshold.SumMethod,
		NotifyDuration: int32(threshold.NotifyDuration),
		IsOn:           threshold.IsOn == 1,
	}}, nil
}
