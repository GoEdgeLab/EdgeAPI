// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// ServerStatBoardChartService 统计看板条目
type ServerStatBoardChartService struct {
	BaseService
}

// EnableServerStatBoardChart 添加图表
func (this *ServerStatBoardChartService) EnableServerStatBoardChart(ctx context.Context, req *pb.EnableServerStatBoardChartRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedServerStatBoardChartDAO.EnableChart(tx, req.ServerStatBoardId, req.MetricChartId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DisableServerStatBoardChart 取消图表
func (this *ServerStatBoardChartService) DisableServerStatBoardChart(ctx context.Context, req *pb.DisableServerStatBoardChartRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedServerStatBoardChartDAO.DisableChart(tx, req.ServerStatBoardId, req.MetricChartId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledServerStatBoardCharts 读取看板中的图表
func (this *ServerStatBoardChartService) FindAllEnabledServerStatBoardCharts(ctx context.Context, req *pb.FindAllEnabledServerStatBoardChartsRequest) (*pb.FindAllEnabledServerStatBoardChartsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	charts, err := models.SharedServerStatBoardChartDAO.FindAllEnabledCharts(tx, req.ServerStatBoardId)
	if err != nil {
		return nil, err
	}
	var pbCharts []*pb.ServerStatBoardChart
	for _, chart := range charts {
		// 指标图表
		metricChart, err := models.SharedMetricChartDAO.FindEnabledMetricChart(tx, int64(chart.ChartId))
		if err != nil {
			return nil, err
		}
		if metricChart == nil {
			continue
		}

		// 指标
		metricItem, err := models.SharedMetricItemDAO.FindEnabledMetricItem(tx, int64(chart.ItemId))
		if err != nil {
			return nil, err
		}
		if metricItem == nil {
			continue
		}

		pbCharts = append(pbCharts, &pb.ServerStatBoardChart{
			Id: int64(chart.Id),
			MetricChart: &pb.MetricChart{
				Id:         int64(metricChart.Id),
				Name:       metricChart.Name,
				Type:       metricChart.Type,
				WidthDiv:   types.Int32(metricChart.WidthDiv),
				ParamsJSON: nil,
				IsOn:       metricChart.IsOn,
				MaxItems:   types.Int32(metricChart.MaxItems),
				MetricItem: &pb.MetricItem{
					Id:         int64(metricItem.Id),
					IsOn:       metricItem.IsOn,
					Code:       metricItem.Code,
					Category:   metricItem.Category,
					Name:       metricItem.Name,
					Keys:       metricItem.DecodeKeys(),
					Period:     types.Int32(metricItem.Period),
					PeriodUnit: metricItem.PeriodUnit,
					Value:      metricItem.Value,
				},
			},
		})
	}

	return &pb.FindAllEnabledServerStatBoardChartsResponse{ServerStatBoardCharts: pbCharts}, nil
}
