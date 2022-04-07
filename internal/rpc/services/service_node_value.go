// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

type NodeValueService struct {
	BaseService
}

// CreateNodeValue 记录数据
func (this *NodeValueService) CreateNodeValue(ctx context.Context, req *pb.CreateNodeValueRequest) (*pb.RPCSuccess, error) {
	role, nodeId, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode, rpcutils.UserTypeDNS, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	var clusterId int64
	switch role {
	case rpcutils.UserTypeNode:
		clusterId, err = models.SharedNodeDAO.FindNodeClusterId(tx, nodeId)
	case rpcutils.UserTypeDNS:
		clusterId, err = models.SharedNSNodeDAO.FindNodeClusterId(tx, nodeId)
	case rpcutils.UserTypeUser:
	}
	if err != nil {
		return nil, err
	}

	err = models.SharedNodeValueDAO.CreateValue(tx, clusterId, role, nodeId, req.Item, req.ValueJSON, req.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 触发节点阈值
	err = models.SharedNodeThresholdDAO.FireNodeThreshold(tx, role, nodeId, req.Item)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ListNodeValues 读取数据
func (this *NodeValueService) ListNodeValues(ctx context.Context, req *pb.ListNodeValuesRequest) (*pb.ListNodeValuesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	values, err := models.SharedNodeValueDAO.ListValues(tx, req.Role, req.NodeId, req.Item, req.Range)
	if err != nil {
		return nil, err
	}
	var pbValues = []*pb.NodeValue{}
	for _, value := range values {
		pbValues = append(pbValues, &pb.NodeValue{
			ValueJSON: value.Value,
			CreatedAt: int64(value.CreatedAt),
		})
	}

	return &pb.ListNodeValuesResponse{NodeValues: pbValues}, nil
}

// SumAllNodeValueStats 读取所有节点的最新数据
func (this *NodeValueService) SumAllNodeValueStats(ctx context.Context, req *pb.SumAllNodeValueStatsRequest) (*pb.SumAllNodeValueStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	var result = &pb.SumAllNodeValueStatsResponse{}

	// traffic
	{
		total, _, _, err := models.SharedNodeValueDAO.SumAllNodeValues(tx, nodeconfigs.NodeRoleNode, nodeconfigs.NodeValueItemTrafficOut, "total", 1, nodeconfigs.NodeValueDurationUnitMinute)
		if err != nil {
			return nil, err
		}
		result.TotalTrafficBytesPerSecond = types.Int64(total) / 60
	}

	// cpu
	{
		_, avg, max, err := models.SharedNodeValueDAO.SumAllNodeValues(tx, nodeconfigs.NodeRoleNode, nodeconfigs.NodeValueItemCPU, "usage", 1, nodeconfigs.NodeValueDurationUnitMinute)
		if err != nil {
			return nil, err
		}
		result.AvgCPUUsage = types.Float32(avg)
		result.MaxCPUUsage = types.Float32(max)
	}

	{
		total, _, _, err := models.SharedNodeValueDAO.SumAllNodeValues(tx, nodeconfigs.NodeRoleNode, nodeconfigs.NodeValueItemCPU, "cores", 1, nodeconfigs.NodeValueDurationUnitMinute)
		if err != nil {
			return nil, err
		}
		result.TotalCPUCores = types.Int32(total)
	}

	// memory
	{
		_, avg, max, err := models.SharedNodeValueDAO.SumAllNodeValues(tx, nodeconfigs.NodeRoleNode, nodeconfigs.NodeValueItemMemory, "usage", 1, nodeconfigs.NodeValueDurationUnitMinute)
		if err != nil {
			return nil, err
		}
		result.AvgMemoryUsage = types.Float32(avg)
		result.MaxMemoryUsage = types.Float32(max)
	}

	{
		total, _, _, err := models.SharedNodeValueDAO.SumAllNodeValues(tx, nodeconfigs.NodeRoleNode, nodeconfigs.NodeValueItemMemory, "total", 1, nodeconfigs.NodeValueDurationUnitMinute)
		if err != nil {
			return nil, err
		}
		result.TotalMemoryBytes = types.Int64(total)
	}

	// load
	{
		_, avg, max, err := models.SharedNodeValueDAO.SumAllNodeValues(tx, nodeconfigs.NodeRoleNode, nodeconfigs.NodeValueItemLoad, "load1m", 1, nodeconfigs.NodeValueDurationUnitMinute)
		if err != nil {
			return nil, err
		}
		result.AvgLoad1Min = types.Float32(avg)
		result.MaxLoad1Min = types.Float32(max)
	}

	{
		_, avg, _, err := models.SharedNodeValueDAO.SumAllNodeValues(tx, nodeconfigs.NodeRoleNode, nodeconfigs.NodeValueItemLoad, "load5m", 1, nodeconfigs.NodeValueDurationUnitMinute)
		if err != nil {
			return nil, err
		}
		result.AvgLoad5Min = types.Float32(avg)
	}

	return result, nil
}
