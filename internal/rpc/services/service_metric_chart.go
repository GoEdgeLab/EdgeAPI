// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

// MetricChartService 指标图表相关服务
type MetricChartService struct {
	BaseService
}

// CreateMetricChart 创建图表
func (this *MetricChartService) CreateMetricChart(ctx context.Context, req *pb.CreateMetricChartRequest) (*pb.CreateMetricChartResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var params = maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}
	chartId, err := models.SharedMetricChartDAO.CreateChart(tx, req.MetricItemId, req.Name, req.Type, req.WidthDiv, req.MaxItems, params, req.IgnoreEmptyKeys, req.IgnoredKeys)
	if err != nil {
		return nil, err
	}
	return &pb.CreateMetricChartResponse{MetricChartId: chartId}, nil
}

// UpdateMetricChart 修改图表
func (this *MetricChartService) UpdateMetricChart(ctx context.Context, req *pb.UpdateMetricChartRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var params = maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}
	err = models.SharedMetricChartDAO.UpdateChart(tx, req.MetricChartId, req.Name, req.Type, req.WidthDiv, req.MaxItems, params, req.IgnoreEmptyKeys, req.IgnoredKeys, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledMetricChart 查找单个图表
func (this *MetricChartService) FindEnabledMetricChart(ctx context.Context, req *pb.FindEnabledMetricChartRequest) (*pb.FindEnabledMetricChartResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	chart, err := models.SharedMetricChartDAO.FindEnabledMetricChart(tx, req.MetricChartId)
	if err != nil {
		return nil, err
	}
	if chart == nil {
		return &pb.FindEnabledMetricChartResponse{MetricChart: nil}, nil
	}
	return &pb.FindEnabledMetricChartResponse{
		MetricChart: &pb.MetricChart{
			Id:              int64(chart.Id),
			Name:            chart.Name,
			Type:            chart.Type,
			WidthDiv:        types.Int32(chart.WidthDiv),
			MaxItems:        types.Int32(chart.MaxItems),
			ParamsJSON:      chart.Params,
			IgnoreEmptyKeys: chart.IgnoreEmptyKeys == 1,
			IgnoredKeys:     chart.DecodeIgnoredKeys(),
			IsOn:            chart.IsOn,
			MetricItem:      &pb.MetricItem{Id: int64(chart.ItemId)},
		},
	}, nil
}

// CountEnabledMetricCharts 计算图表数量
func (this *MetricChartService) CountEnabledMetricCharts(ctx context.Context, req *pb.CountEnabledMetricChartsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedMetricChartDAO.CountEnabledCharts(tx, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledMetricCharts 列出单页图表
func (this *MetricChartService) ListEnabledMetricCharts(ctx context.Context, req *pb.ListEnabledMetricChartsRequest) (*pb.ListEnabledMetricChartsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	charts, err := models.SharedMetricChartDAO.ListEnabledCharts(tx, req.MetricItemId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbCharts []*pb.MetricChart
	for _, chart := range charts {
		pbCharts = append(pbCharts, &pb.MetricChart{
			Id:              int64(chart.Id),
			Name:            chart.Name,
			Type:            chart.Type,
			WidthDiv:        types.Int32(chart.WidthDiv),
			MaxItems:        types.Int32(chart.MaxItems),
			ParamsJSON:      chart.Params,
			IgnoreEmptyKeys: chart.IgnoreEmptyKeys == 1,
			IgnoredKeys:     chart.DecodeIgnoredKeys(),
			IsOn:            chart.IsOn,
		})
	}
	return &pb.ListEnabledMetricChartsResponse{MetricCharts: pbCharts}, nil
}

// DeleteMetricChart 删除图表
func (this *MetricChartService) DeleteMetricChart(ctx context.Context, req *pb.DeleteMetricChartRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMetricChartDAO.DisableMetricChart(tx, req.MetricChartId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
