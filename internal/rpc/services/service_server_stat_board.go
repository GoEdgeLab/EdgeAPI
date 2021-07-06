// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
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
	topNodeStats, err := stats.SharedNodeTrafficHourlyStatDAO.FindTopNodeStatsWithClusterId(tx, "node", req.NodeClusterId, hourFrom, hourTo)
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
	cpuValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, "node", req.NodeClusterId, nodeconfigs.NodeValueItemCPU, "usage", nodeconfigs.NodeValueRangeMinute)
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

	memoryValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, "node", req.NodeClusterId, nodeconfigs.NodeValueItemMemory, "usage", nodeconfigs.NodeValueRangeMinute)
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

	loadValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, "node", req.NodeClusterId, nodeconfigs.NodeValueItemLoad, "load5m", nodeconfigs.NodeValueRangeMinute)
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

// ComposeServerStatNodeBoard 组合节点看板数据
func (this *ServerStatBoardService) ComposeServerStatNodeBoard(ctx context.Context, req *pb.ComposeServerStatNodeBoardRequest) (*pb.ComposeServerStatNodeBoardResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var result = &pb.ComposeServerStatNodeBoardResponse{}

	// 在线状态
	var tx = this.NullTx()
	node, err := models.SharedNodeDAO.FindEnabledNode(tx, req.NodeId)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, errors.New("node not found")
	}

	status, err := node.DecodeStatus()
	if err != nil {
		return nil, err
	}
	if status != nil && time.Now().Unix()-status.UpdatedAt < 60 {
		result.IsActive = true
		result.CacheDiskSize = status.CacheTotalDiskSize
		result.CacheMemorySize = status.CacheTotalMemorySize

	}

	// 流量
	{
		value, err := models.SharedNodeValueDAO.FindLatestNodeValue(tx, "node", int64(node.Id), nodeconfigs.NodeValueItemTrafficIn)
		if err != nil {
			return nil, err
		}
		if value != nil && time.Now().Unix()-int64(value.CreatedAt) < 120 {
			result.TrafficInBytes = value.DecodeMapValue().GetInt64("total")
		}
	}
	{
		value, err := models.SharedNodeValueDAO.FindLatestNodeValue(tx, "node", int64(node.Id), nodeconfigs.NodeValueItemTrafficOut)
		if err != nil {
			return nil, err
		}
		if value != nil && time.Now().Unix()-int64(value.CreatedAt) < 120 {
			result.TrafficOutBytes = value.DecodeMapValue().GetInt64("total")
		}
	}

	// 连接数
	{
		value, err := models.SharedNodeValueDAO.FindLatestNodeValue(tx, "node", int64(node.Id), nodeconfigs.NodeValueItemConnections)
		if err != nil {
			return nil, err
		}
		if value != nil && time.Now().Unix()-int64(value.CreatedAt) < 120 {
			result.CountConnections = value.DecodeMapValue().GetInt64("total")
		}
	}

	// 请求量
	{
		value, err := models.SharedNodeValueDAO.FindLatestNodeValue(tx, "node", int64(node.Id), nodeconfigs.NodeValueItemRequests)
		if err != nil {
			return nil, err
		}
		if value != nil && time.Now().Unix()-int64(value.CreatedAt) < 120 {
			result.CountRequests = value.DecodeMapValue().GetInt64("total")
		}
	}
	{
		value, err := models.SharedNodeValueDAO.FindLatestNodeValue(tx, "node", int64(node.Id), nodeconfigs.NodeValueItemAttackRequests)
		if err != nil {
			return nil, err
		}
		if value != nil && time.Now().Unix()-int64(value.CreatedAt) < 120 {
			result.CountAttackRequests = value.DecodeMapValue().GetInt64("total")
		}
	}

	// CPU
	{
		value, err := models.SharedNodeValueDAO.FindLatestNodeValue(tx, "node", int64(node.Id), nodeconfigs.NodeValueItemCPU)
		if err != nil {
			return nil, err
		}
		if value != nil && time.Now().Unix()-int64(value.CreatedAt) < 120 {
			result.CpuUsage = value.DecodeMapValue().GetFloat32("usage")
		}
	}

	// 内存
	{
		value, err := models.SharedNodeValueDAO.FindLatestNodeValue(tx, "node", int64(node.Id), nodeconfigs.NodeValueItemMemory)
		if err != nil {
			return nil, err
		}
		if value != nil && time.Now().Unix()-int64(value.CreatedAt) < 120 {
			m := value.DecodeMapValue()
			result.MemoryUsage = m.GetFloat32("usage")
			result.MemoryTotalSize = m.GetInt64("total")
		}
	}

	// 负载
	{
		value, err := models.SharedNodeValueDAO.FindLatestNodeValue(tx, "node", int64(node.Id), nodeconfigs.NodeValueItemLoad)
		if err != nil {
			return nil, err
		}
		if value != nil && time.Now().Unix()-int64(value.CreatedAt) < 120 {
			result.Load = value.DecodeMapValue().GetFloat32("load1m")
		}
	}

	// 按日流量统计
	dayFrom := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dailyTrafficStats, err := stats.SharedNodeTrafficDailyStatDAO.FindDailyStats(tx, "node", req.NodeId, dayFrom, timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	for _, stat := range dailyTrafficStats {
		result.DailyTrafficStats = append(result.DailyTrafficStats, &pb.ComposeServerStatNodeBoardResponse_DailyTrafficStat{
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
	hourlyTrafficStats, err := stats.SharedNodeTrafficHourlyStatDAO.FindHourlyStatsWithNodeId(tx, "node", req.NodeId, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range hourlyTrafficStats {
		result.HourlyTrafficStats = append(result.HourlyTrafficStats, &pb.ComposeServerStatNodeBoardResponse_HourlyTrafficStat{
			Hour:                stat.Hour,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
		})
	}

	// 域名排行
	topDomainStats, err := stats.SharedServerDomainHourlyStatDAO.FindTopDomainStatsWithNodeId(tx, req.NodeId, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range topDomainStats {
		result.TopDomainStats = append(result.TopDomainStats, &pb.ComposeServerStatNodeBoardResponse_DomainStat{
			ServerId:      int64(stat.ServerId),
			Domain:        stat.Domain,
			CountRequests: int64(stat.CountRequests),
			Bytes:         int64(stat.Bytes),
		})
	}

	// CPU、内存、负载
	cpuValues, err := models.SharedNodeValueDAO.ListValues(tx, "node", req.NodeId, nodeconfigs.NodeValueItemCPU, nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range cpuValues {
		valueJSON, err := json.Marshal(types.Float32(v.DecodeMapValue().GetFloat32("usage")))
		if err != nil {
			return nil, err
		}
		result.CpuNodeValues = append(result.CpuNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	memoryValues, err := models.SharedNodeValueDAO.ListValues(tx, "node", req.NodeId, nodeconfigs.NodeValueItemMemory, nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range memoryValues {
		valueJSON, err := json.Marshal(types.Float32(v.DecodeMapValue().GetFloat32("usage")))
		if err != nil {
			return nil, err
		}
		result.MemoryNodeValues = append(result.MemoryNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	loadValues, err := models.SharedNodeValueDAO.ListValues(tx, "node", req.NodeId, nodeconfigs.NodeValueItemLoad, nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range loadValues {
		valueJSON, err := json.Marshal(types.Float32(v.DecodeMapValue().GetFloat32("load5m")))
		if err != nil {
			return nil, err
		}
		result.LoadNodeValues = append(result.LoadNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	// 指标
	var clusterId = int64(node.ClusterId)
	clusterMetricItems, err := models.SharedNodeClusterMetricItemDAO.FindAllClusterItems(tx, clusterId, serverconfigs.MetricItemCategoryHTTP)
	if err != nil {
		return nil, err
	}
	var pbMetricCharts = []*pb.ComposeServerStatNodeBoardResponse_MetricData{}
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
				itemStats, err := models.SharedMetricStatDAO.FindLatestItemStatsWithNodeId(tx, req.NodeId, itemId, types.Int32(item.Version), 10)
				if err != nil {
					return nil, err
				}

				for _, stat := range itemStats {
					// 当前时间总和
					count, total, err := models.SharedMetricSumStatDAO.FindNodeSum(tx, req.NodeId, stat.Time, itemId, types.Int32(item.Version))
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
				itemStats, err := models.SharedMetricStatDAO.FindItemStatsWithNodeIdAndLastTime(tx, req.NodeId, itemId, types.Int32(item.Version), 10)
				if err != nil {
					return nil, err
				}
				for _, stat := range itemStats {
					// 当前时间总和
					// 当前时间总和
					count, total, err := models.SharedMetricSumStatDAO.FindNodeSum(tx, req.NodeId, stat.Time, itemId, types.Int32(item.Version))
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
			pbMetricCharts = append(pbMetricCharts, &pb.ComposeServerStatNodeBoardResponse_MetricData{
				MetricChart: pbChart,
				MetricStats: pbStats,
			})
		}
	}
	result.MetricCharts = pbMetricCharts

	return result, nil
}
