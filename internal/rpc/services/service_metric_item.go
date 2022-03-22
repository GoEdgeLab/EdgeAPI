// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// MetricItemService 指标相关服务
type MetricItemService struct {
	BaseService
}

// CreateMetricItem 创建指标
func (this *MetricItemService) CreateMetricItem(ctx context.Context, req *pb.CreateMetricItemRequest) (*pb.CreateMetricItemResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	itemId, err := models.SharedMetricItemDAO.CreateItem(tx, req.Code, req.Category, req.Name, req.Keys, req.Period, req.PeriodUnit, req.Value, req.IsPublic)
	if err != nil {
		return nil, err
	}
	return &pb.CreateMetricItemResponse{MetricItemId: itemId}, nil
}

// UpdateMetricItem 修改指标
func (this *MetricItemService) UpdateMetricItem(ctx context.Context, req *pb.UpdateMetricItemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMetricItemDAO.UpdateItem(tx, req.MetricItemId, req.Name, req.Keys, req.Period, req.PeriodUnit, req.Value, req.IsOn, req.IsPublic)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledMetricItem 查找单个指标信息
func (this *MetricItemService) FindEnabledMetricItem(ctx context.Context, req *pb.FindEnabledMetricItemRequest) (*pb.FindEnabledMetricItemResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	item, err := models.SharedMetricItemDAO.FindEnabledMetricItem(tx, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &pb.FindEnabledMetricItemResponse{MetricItem: nil}, nil
	}
	return &pb.FindEnabledMetricItemResponse{MetricItem: &pb.MetricItem{
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
	}}, nil
}

// CountAllEnabledMetricItems 计算指标数量
func (this *MetricItemService) CountAllEnabledMetricItems(ctx context.Context, req *pb.CountAllEnabledMetricItemsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedMetricItemDAO.CountEnabledItems(tx, req.Category)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledMetricItems 列出单页指标
func (this *MetricItemService) ListEnabledMetricItems(ctx context.Context, req *pb.ListEnabledMetricItemsRequest) (*pb.ListEnabledMetricItemsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	items, err := models.SharedMetricItemDAO.ListEnabledItems(tx, req.Category, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbItems = []*pb.MetricItem{}
	for _, item := range items {
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

	return &pb.ListEnabledMetricItemsResponse{MetricItems: pbItems}, nil
}

// DeleteMetricItem 删除指标
func (this *MetricItemService) DeleteMetricItem(ctx context.Context, req *pb.DeleteMetricItemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMetricItemDAO.DisableMetricItem(tx, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
