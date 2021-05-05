// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type NodeValueService struct {
	BaseService
}

// CreateNodeValue 记录数据
func (this *NodeValueService) CreateNodeValue(ctx context.Context, req *pb.CreateNodeValueRequest) (*pb.RPCSuccess, error) {
	role, nodeId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeValueDAO.CreateValue(tx, role, nodeId, req.Item, req.ValueJSON, req.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 触发阈值
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
	pbValues := []*pb.NodeValue{}
	for _, value := range values {
		pbValues = append(pbValues, &pb.NodeValue{
			ValueJSON: []byte(value.Value),
			CreatedAt: int64(value.CreatedAt),
		})
	}

	return &pb.ListNodeValuesResponse{NodeValues: pbValues}, nil
}
