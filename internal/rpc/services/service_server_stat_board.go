// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// ServerStatBoardService 统计看板条目
type ServerStatBoardService struct {
	BaseService
}

// FindAllEnabledServerStatBoards 读取所有看板
func (this *ServerStatBoardService) FindAllEnabledServerStatBoards(ctx context.Context, req *pb.FindAllEnabledServerStatBoardsRequest) (*pb.FindAllEnabledServerStatBoardsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	boards, err := models.SharedServerStatBoardDAO.FindAllEnabledBoards(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	var pbBoards = []*pb.ServerStatBoard{}
	for _, board := range boards {
		pbBoards = append(pbBoards, &pb.ServerStatBoard{
			Id:   int64(board.Id),
			Name: board.Name,
			IsOn: board.IsOn == 1,
		})
	}

	return &pb.FindAllEnabledServerStatBoardsResponse{
		ServerStatBoards: pbBoards,
	}, nil
}

// ComposeServerStatNodeClusterBoard 组合看板数据
func (this *ServerStatBoardService) ComposeServerStatNodeClusterBoard(ctx context.Context, req *pb.ComposeServerStatNodeClusterBoardRequest) (*pb.ComposeServerStatNodeClusterBoardResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var result = &pb.ComposeServerStatNodeClusterBoardResponse{}

	// 统计数字
	countActiveNodes, err := models.SharedNodeDAO.CountAllEnabledNodesMatch(tx, req.NodeClusterId, configutils.BoolStateAll, configutils.BoolStateYes, "", 0, 0)
	if err != nil {
		return nil, err
	}
	result.CountActiveNodes = countActiveNodes

	countInactiveNodes, err := models.SharedNodeDAO.CountAllEnabledNodesMatch(tx, req.NodeClusterId, configutils.BoolStateAll, configutils.BoolStateNo, "", 0, 0)
	if err != nil {
		return nil, err
	}
	result.CountInactiveNodes = countInactiveNodes

	countUsers, err := models.SharedUserDAO.CountAllEnabledUsers(tx, req.NodeClusterId, "")
	if err != nil {
		return nil, err
	}
	result.CountUsers = countUsers

	countServers, err := models.SharedServerDAO.CountAllEnabledServersWithNodeClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result.CountServers = countServers

	// 按日流量统计
	dayFrom := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dailyTrafficStats, err := stats.SharedNodeClusterTrafficDailyStatDAO.FindDailyStats(tx, req.NodeClusterId, dayFrom, timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	for _, stat := range dailyTrafficStats {
		result.DailyTrafficStats = append(result.DailyTrafficStats, &pb.ComposeServerStatNodeClusterBoardResponse_DailyTrafficStat{
			Day:                 stat.Day,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
		})
	}

	// 小时流量统计
	hourFrom := timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	hourTo := timeutil.Format("YmdH")
	hourlyTrafficStats, err := stats.SharedNodeTrafficHourlyStatDAO.FindHourlyStatsWithClusterId(tx, req.NodeClusterId, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range hourlyTrafficStats {
		result.HourlyTrafficStats = append(result.HourlyTrafficStats, &pb.ComposeServerStatNodeClusterBoardResponse_HourlyTrafficStat{
			Hour:                stat.Hour,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
		})
	}

	// 节点排行
	topNodeStats, err := stats.SharedNodeTrafficHourlyStatDAO.FindTopNodeStatsWithClusterId(tx, req.NodeClusterId, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range topNodeStats {
		nodeName, err := models.SharedNodeDAO.FindNodeName(tx, int64(stat.NodeId))
		if err != nil {
			return nil, err
		}
		if len(nodeName) == 0 {
			continue
		}
		result.TopNodeStats = append(result.TopNodeStats, &pb.ComposeServerStatNodeClusterBoardResponse_NodeStat{
			NodeId:        int64(stat.NodeId),
			NodeName:      nodeName,
			CountRequests: int64(stat.CountRequests),
			Bytes:         int64(stat.Bytes),
		})
	}

	// 域名排行
	topDomainStats, err := stats.SharedServerDomainHourlyStatDAO.FindTopDomainStatsWithClusterId(tx, req.NodeClusterId, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range topDomainStats {
		result.TopDomainStats = append(result.TopDomainStats, &pb.ComposeServerStatNodeClusterBoardResponse_DomainStat{
			ServerId:      int64(stat.ServerId),
			Domain:        stat.Domain,
			CountRequests: int64(stat.CountRequests),
			Bytes:         int64(stat.Bytes),
		})
	}

	// CPU、内存、负载
	cpuValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, req.NodeClusterId, "node", nodeconfigs.NodeValueItemCPU, "usage", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range cpuValues {
		valueJSON, err := json.Marshal(types.Float32(v.Value))
		if err != nil {
			return nil, err
		}
		result.CpuNodeValues = append(result.CpuNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	memoryValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, req.NodeClusterId, "node", nodeconfigs.NodeValueItemMemory, "usage", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range memoryValues {
		valueJSON, err := json.Marshal(types.Float32(v.Value))
		if err != nil {
			return nil, err
		}
		result.MemoryNodeValues = append(result.MemoryNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	loadValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, req.NodeClusterId, "node", nodeconfigs.NodeValueItemLoad, "load5m", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range loadValues {
		valueJSON, err := json.Marshal(types.Float32(v.Value))
		if err != nil {
			return nil, err
		}
		result.LoadNodeValues = append(result.LoadNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	// 指标
	clusterMetricItems, err := models.SharedNodeClusterMetricItemDAO.FindAllClusterItems(tx, req.NodeClusterId, serverconfigs.MetricItemCategoryHTTP)
	if err != nil {
		return nil, err
	}
	var pbMetricCharts = []*pb.ComposeServerStatNodeClusterBoardResponse_MetricData{}
	for _, clusterMetricItem := range clusterMetricItems {
		if clusterMetricItem.IsOn != 1 {
			continue
		}
		var itemId = int64(clusterMetricItem.ItemId)
		charts, err := models.SharedMetricChartDAO.FindAllEnabledCharts(tx, itemId)
		if err != nil {
			return nil, err
		}

		item, err := models.SharedMetricItemDAO.FindEnabledMetricItem(tx, itemId)
		if err != nil {
			return nil, err
		}
		if item == nil || item.IsOn == 0 {
			continue
		}

		for _, chart := range charts {
			if chart.IsOn == 0 {
				continue
			}

			var pbChart = &pb.MetricChart{
				Id:         int64(chart.Id),
				Name:       chart.Name,
				Type:       chart.Type,
				WidthDiv:   chart.WidthDiv,
				ParamsJSON: nil,
				IsOn:       chart.IsOn == 1,
				MaxItems:   types.Int32(chart.MaxItems),
				MetricItem: &pb.MetricItem{
					Id:         itemId,
					PeriodUnit: item.PeriodUnit,
					Period:     types.Int32(item.Period),
					Name:       item.Name,
					Value:      item.Value,
					Category:   item.Category,
					Keys:       item.DecodeKeys(),
					Code:       item.Code,
					IsOn:       item.IsOn == 1,
				},
			}
			var pbStats = []*pb.MetricStat{}
			switch chart.Type {
			case serverconfigs.MetricChartTypeTimeLine:
				itemStats, err := models.SharedMetricStatDAO.FindLatestItemStatsWithClusterId(tx, req.NodeClusterId, itemId, types.Int32(item.Version), 10)
				if err != nil {
					return nil, err
				}

				for _, stat := range itemStats {
					// 当前时间总和
					count, total, err := models.SharedMetricSumStatDAO.FindClusterSum(tx, req.NodeClusterId, stat.Time, itemId, types.Int32(item.Version))
					if err != nil {
						return nil, err
					}

					pbStats = append(pbStats, &pb.MetricStat{
						Id:          int64(stat.Id),
						Hash:        stat.Hash,
						ServerId:    0,
						ItemId:      0,
						Keys:        stat.DecodeKeys(),
						Value:       types.Float32(stat.Value),
						Time:        stat.Time,
						Version:     0,
						NodeCluster: nil,
						Node:        nil,
						Server:      nil,
						SumCount:    count,
						SumTotal:    total,
					})
				}
			default:
				itemStats, err := models.SharedMetricStatDAO.FindItemStatsWithClusterIdAndLastTime(tx, req.NodeClusterId, itemId, types.Int32(item.Version), 10)
				if err != nil {
					return nil, err
				}
				for _, stat := range itemStats {
					// 当前时间总和
					// 当前时间总和
					count, total, err := models.SharedMetricSumStatDAO.FindClusterSum(tx, req.NodeClusterId, stat.Time, itemId, types.Int32(item.Version))
					if err != nil {
						return nil, err
					}

					pbStats = append(pbStats, &pb.MetricStat{
						Id:          int64(stat.Id),
						Hash:        stat.Hash,
						ServerId:    0,
						ItemId:      0,
						Keys:        stat.DecodeKeys(),
						Value:       types.Float32(stat.Value),
						Time:        stat.Time,
						Version:     0,
						NodeCluster: nil,
						Node:        nil,
						Server:      nil,
						SumCount:    count,
						SumTotal:    total,
					})
				}
			}
			pbMetricCharts = append(pbMetricCharts, &pb.ComposeServerStatNodeClusterBoardResponse_MetricData{
				MetricChart: pbChart,
				MetricStats: pbStats,
			})
		}
	}
	result.MetricCharts = pbMetricCharts

	return result, nil
}
