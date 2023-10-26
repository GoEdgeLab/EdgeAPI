// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/tasks"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/configutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// ServerStatBoardService 统计看板条目
type ServerStatBoardService struct {
	BaseService
}

// FindAllEnabledServerStatBoards 读取所有看板
func (this *ServerStatBoardService) FindAllEnabledServerStatBoards(ctx context.Context, req *pb.FindAllEnabledServerStatBoardsRequest) (*pb.FindAllEnabledServerStatBoardsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
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
			IsOn: board.IsOn,
		})
	}

	return &pb.FindAllEnabledServerStatBoardsResponse{
		ServerStatBoards: pbBoards,
	}, nil
}

// ComposeServerStatNodeClusterBoard 组合看板数据
func (this *ServerStatBoardService) ComposeServerStatNodeClusterBoard(ctx context.Context, req *pb.ComposeServerStatNodeClusterBoardRequest) (*pb.ComposeServerStatNodeClusterBoardResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var result = &pb.ComposeServerStatNodeClusterBoardResponse{}

	// 统计数字
	countActiveNodes, err := models.SharedNodeDAO.CountAllEnabledNodesMatch(tx, req.NodeClusterId, configutils.BoolStateAll, configutils.BoolStateYes, "", 0, 0, 0, true)
	if err != nil {
		return nil, err
	}
	result.CountActiveNodes = countActiveNodes

	countInactiveNodes, err := models.SharedNodeDAO.CountAllEnabledNodesMatch(tx, req.NodeClusterId, configutils.BoolStateAll, configutils.BoolStateNo, "", 0, 0, 0, true)
	if err != nil {
		return nil, err
	}
	result.CountInactiveNodes = countInactiveNodes

	countUsers, err := models.SharedUserDAO.CountAllEnabledUsers(tx, req.NodeClusterId, "", false)
	if err != nil {
		return nil, err
	}
	result.CountUsers = countUsers

	countServers, err := models.SharedServerDAO.CountAllEnabledServersWithNodeClusterId(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	result.CountServers = countServers

	// 当月总流量
	monthlyTrafficStat, err := stats.SharedNodeClusterTrafficDailyStatDAO.SumDailyStat(tx, req.NodeClusterId, timeutil.Format("Ym01"), timeutil.Format("Ym31"))
	if err != nil {
		return nil, err
	}
	if monthlyTrafficStat != nil {
		result.MonthlyTrafficBytes = int64(monthlyTrafficStat.Bytes)
	}

	// 按日流量统计
	var dayFrom = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dailyTrafficStats, err := stats.SharedNodeClusterTrafficDailyStatDAO.FindDailyStats(tx, req.NodeClusterId, dayFrom, timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	var dailyTrafficBytes int64
	var lastDailyTrafficBytes int64
	for _, stat := range dailyTrafficStats {
		if stat.Day == timeutil.Format("Ymd") { // 今天
			dailyTrafficBytes = int64(stat.Bytes)
		} else if stat.Day == timeutil.Format("Ymd", time.Now().AddDate(0, 0, -1)) {
			lastDailyTrafficBytes = int64(stat.Bytes)
		}

		result.DailyTrafficStats = append(result.DailyTrafficStats, &pb.ComposeServerStatNodeClusterBoardResponse_DailyTrafficStat{
			Day:                 stat.Day,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}
	result.DailyTrafficBytes = dailyTrafficBytes
	result.LastDailyTrafficBytes = lastDailyTrafficBytes

	// 小时流量统计
	var hourFrom = timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	var hourTo = timeutil.Format("YmdH")
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
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}

	// 节点排行
	topNodeStats, err := stats.SharedNodeTrafficHourlyStatDAO.FindTopNodeStatsWithClusterId(tx, "node", req.NodeClusterId, hourFrom, hourTo, 10)
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
			NodeId:              int64(stat.NodeId),
			NodeName:            nodeName,
			CountRequests:       int64(stat.CountRequests),
			Bytes:               int64(stat.Bytes),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}

	// CPU、内存、负载
	cpuValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, "node", req.NodeClusterId, nodeconfigs.NodeValueItemCPU, "usage", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range cpuValues {
		result.CpuNodeValues = append(result.CpuNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	memoryValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, "node", req.NodeClusterId, nodeconfigs.NodeValueItemMemory, "usage", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range memoryValues {
		result.MemoryNodeValues = append(result.MemoryNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	loadValues, err := models.SharedNodeValueDAO.ListValuesWithClusterId(tx, "node", req.NodeClusterId, nodeconfigs.NodeValueItemLoad, "load1m", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range loadValues {
		result.LoadNodeValues = append(result.LoadNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	var pbCharts []*pb.MetricDataChart
	charts, ok := tasks.SharedCacheTaskManager.GetCluster(tasks.CacheKeyFindNodeClusterMetricDataCharts, req.NodeClusterId, serverconfigs.MetricItemCategoryHTTP)
	if ok {
		pbCharts = charts.([]*pb.MetricDataChart)
	}
	result.MetricDataCharts = pbCharts

	return result, nil
}

// ComposeServerStatNodeBoard 组合节点看板数据
func (this *ServerStatBoardService) ComposeServerStatNodeBoard(ctx context.Context, req *pb.ComposeServerStatNodeBoardRequest) (*pb.ComposeServerStatNodeBoardResponse, error) {
	_, err := this.ValidateAdmin(ctx)
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

	// 当月总流量
	monthlyTrafficStat, err := models.SharedNodeTrafficDailyStatDAO.SumDailyStat(tx, nodeconfigs.NodeRoleNode, req.NodeId, timeutil.Format("Ym01"), timeutil.Format("Ym31"))
	if err != nil {
		return nil, err
	}
	if monthlyTrafficStat != nil {
		result.MonthlyTrafficBytes = int64(monthlyTrafficStat.Bytes)
	}

	// 按日流量统计
	var dayFrom = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dailyTrafficStats, err := models.SharedNodeTrafficDailyStatDAO.FindDailyStats(tx, "node", req.NodeId, dayFrom, timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	var dailyTrafficBytes int64
	var lastDailyTrafficBytes int64
	for _, stat := range dailyTrafficStats {
		if stat.Day == timeutil.Format("Ymd") { // 当天
			dailyTrafficBytes = int64(stat.Bytes)
		} else if stat.Day == timeutil.Format("Ymd", time.Now().AddDate(0, 0, -1)) { // 昨天
			lastDailyTrafficBytes = int64(stat.Bytes)
		}

		result.DailyTrafficStats = append(result.DailyTrafficStats, &pb.ComposeServerStatNodeBoardResponse_DailyTrafficStat{
			Day:                 stat.Day,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}
	result.DailyTrafficBytes = dailyTrafficBytes
	result.LastDailyTrafficBytes = lastDailyTrafficBytes

	// 小时流量统计
	var hourFrom = timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	var hourTo = timeutil.Format("YmdH")
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
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}

	// CPU、内存、负载
	cpuValues, err := models.SharedNodeValueDAO.ListValues(tx, "node", req.NodeId, nodeconfigs.NodeValueItemCPU, nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range cpuValues {
		result.CpuNodeValues = append(result.CpuNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	memoryValues, err := models.SharedNodeValueDAO.ListValues(tx, "node", req.NodeId, nodeconfigs.NodeValueItemMemory, nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range memoryValues {
		result.MemoryNodeValues = append(result.MemoryNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	loadValues, err := models.SharedNodeValueDAO.ListValues(tx, "node", req.NodeId, nodeconfigs.NodeValueItemLoad, nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range loadValues {
		result.LoadNodeValues = append(result.LoadNodeValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	networkPacketsValues, err := models.SharedNodeValueDAO.ListValues(tx, "node", req.NodeId, nodeconfigs.NodeValueItemNetworkPackets, nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range networkPacketsValues {
		result.NetworkPacketsValues = append(result.NetworkPacketsValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	cacheDirValues, err := models.SharedNodeValueDAO.ListValues(tx, "node", req.NodeId, nodeconfigs.NodeValueItemCacheDir, nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range cacheDirValues {
		result.CacheDirsValues = append(result.CacheDirsValues, &pb.NodeValue{
			ValueJSON: v.Value,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	// 指标
	var clusterId = int64(node.ClusterId)
	var pbCharts []*pb.MetricDataChart
	charts, ok := tasks.SharedCacheTaskManager.GetNode(tasks.CacheKeyFindNodeMetricDataCharts, clusterId, req.NodeId, serverconfigs.MetricItemCategoryHTTP)
	if ok {
		pbCharts = charts.([]*pb.MetricDataChart)
	}
	result.MetricDataCharts = pbCharts

	return result, nil
}

// ComposeServerStatBoard 组合服务看板数据
func (this *ServerStatBoardService) ComposeServerStatBoard(ctx context.Context, req *pb.ComposeServerStatBoardRequest) (*pb.ComposeServerStatBoardResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var result = &pb.ComposeServerStatBoardResponse{}
	var tx = this.NullTx()

	// 用户ID
	userId, err := models.SharedServerDAO.FindServerUserId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	var bandwidthAglo = ""
	if userId > 0 {
		bandwidthAglo, err = models.SharedUserDAO.FindUserBandwidthAlgoForView(tx, userId, nil)
		if err != nil {
			return nil, err
		}
	}

	// 带宽统计
	{
		var month = timeutil.Format("Ym")
		var day = timeutil.Format("Ymd")

		// 当前N分钟区间
		{
			// 查询最近的三个时段，以尽可能获取数据
			var timestamp = time.Now().Unix() / 300 * 300
			var minute1 = timeutil.FormatTime("Hi", timestamp)
			var minute2 = timeutil.FormatTime("Hi", timestamp-300)
			var minute3 = timeutil.FormatTime("Hi", timestamp-300*2)

			for _, minute := range []string{minute1, minute2, minute3} {
				bytes, err := models.SharedServerBandwidthStatDAO.FindMinutelyPeekBandwidthBytes(tx, req.ServerId, day, minute, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
				if err != nil {
					return nil, err
				}

				if bytes > 0 {
					result.MinutelyPeekBandwidthBytes = bytes
					break
				}
			}
		}

		// 当天
		{
			bytes, err := models.SharedServerBandwidthStatDAO.FindDailyPeekBandwidthBytes(tx, req.ServerId, day, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
			if err != nil {
				return nil, err
			}
			result.DailyPeekBandwidthBytes = bytes
		}

		// 当月
		{
			bytes, err := models.SharedServerBandwidthStatDAO.FindMonthlyPeekBandwidthBytes(tx, req.ServerId, month, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
			if err != nil {
				return nil, err
			}
			result.MonthlyPeekBandwidthBytes = bytes
		}

		// 上月
		{
			bytes, err := models.SharedServerBandwidthStatDAO.FindMonthlyPeekBandwidthBytes(tx, req.ServerId, timeutil.Format("Ym", time.Now().AddDate(0, -1, 0)), bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
			if err != nil {
				return nil, err
			}
			result.LastMonthlyPeekBandwidthBytes = bytes
		}
	}

	{
		var bandwidthMinutes = utils.RangeMinutes(time.Now(), 12, 5)
		var bandwidthStatMap = map[string]*pb.ServerBandwidthStat{}
		var timeFrom = ""
		var timeTo = ""
		for _, r := range utils.GroupMinuteRanges(bandwidthMinutes) {
			if len(timeFrom) == 0 || timeFrom > r.Day+r.MinuteFrom {
				timeFrom = r.Day + r.MinuteFrom
			}
			if len(timeTo) == 0 || timeTo < r.Day+r.MinuteTo {
				timeTo = r.Day + r.MinuteTo
			}

			bandwidthStats, err := models.SharedServerBandwidthStatDAO.FindServerStats(tx, req.ServerId, r.Day, r.MinuteFrom, r.MinuteTo, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
			if err != nil {
				return nil, err
			}
			for _, stat := range bandwidthStats {
				bandwidthStatMap[stat.Day+"@"+stat.TimeAt] = &pb.ServerBandwidthStat{
					Id:       int64(stat.Id),
					ServerId: int64(stat.ServerId),
					Day:      stat.Day,
					TimeAt:   stat.TimeAt,
					Bytes:    int64(stat.Bytes),
					Bits:     int64(stat.Bytes * 8),
				}
			}
		}
		var pbBandwidthStats = []*pb.ServerBandwidthStat{}
		for _, minute := range bandwidthMinutes {
			stat, ok := bandwidthStatMap[minute.Day+"@"+minute.Minute]
			if ok {
				pbBandwidthStats = append(pbBandwidthStats, stat)
			} else {
				var bytes = ServerBandwidthGetCacheBytes(req.ServerId, minute.Minute) // 从当前缓存中读取
				pbBandwidthStats = append(pbBandwidthStats, &pb.ServerBandwidthStat{
					Id:       0,
					ServerId: req.ServerId,
					Day:      minute.Day,
					TimeAt:   minute.Minute,
					Bytes:    bytes,
					Bits:     bytes * 8,
				})
			}
		}
		result.MinutelyBandwidthStats = pbBandwidthStats

		// percentile
		if len(timeFrom) > 0 && len(timeTo) > 0 {
			var percentile = systemconfigs.DefaultBandwidthPercentile
			userUIConfig, _ := models.SharedSysSettingDAO.ReadUserUIConfig(tx)
			if userUIConfig != nil && userUIConfig.TrafficStats.BandwidthPercentile > 0 {
				percentile = userUIConfig.TrafficStats.BandwidthPercentile
			}
			result.BandwidthPercentile = percentile

			percentileStat, err := models.SharedServerBandwidthStatDAO.FindPercentileBetweenTimes(tx, req.ServerId, timeFrom, timeTo, percentile, bandwidthAglo == systemconfigs.BandwidthAlgoAvg)
			if err != nil {
				return nil, err
			}
			if percentileStat != nil {
				result.MinutelyNthBandwidthStat = &pb.ServerBandwidthStat{
					Day:    percentileStat.Day,
					TimeAt: percentileStat.TimeAt,
					Bytes:  int64(percentileStat.Bytes),
					Bits:   int64(percentileStat.Bytes * 8),
				}
			}
		}
	}

	// 按日流量统计
	var dayFrom = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dailyTrafficStats, err := models.SharedServerBandwidthStatDAO.FindDailyStats(tx, req.ServerId, dayFrom, timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	for _, stat := range dailyTrafficStats {
		result.DailyTrafficStats = append(result.DailyTrafficStats, &pb.ComposeServerStatBoardResponse_DailyTrafficStat{
			Day:                 stat.Day,
			Bytes:               int64(stat.TotalBytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}

	// 小时流量统计
	var hourFrom = timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	var hourTo = timeutil.Format("YmdH")
	hourlyTrafficStats, err := models.SharedServerBandwidthStatDAO.FindHourlyStats(tx, req.ServerId, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range hourlyTrafficStats {
		result.HourlyTrafficStats = append(result.HourlyTrafficStats, &pb.ComposeServerStatBoardResponse_HourlyTrafficStat{
			Hour:                stat.Day + stat.TimeAt[:2],
			Bytes:               int64(stat.TotalBytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		})
	}

	// 地区流量排行
	totalCountryBytes, err := stats.SharedServerRegionCountryDailyStatDAO.SumDailyTotalBytesWithServerId(tx, timeutil.Format("Ymd"), req.ServerId)
	if err != nil {
		return nil, err
	}

	if totalCountryBytes > 0 {
		topCountryStats, err := stats.SharedServerRegionCountryDailyStatDAO.ListServerStats(tx, req.ServerId, timeutil.Format("Ymd"), "bytes", 0, 100)
		if err != nil {
			return nil, err
		}

		for _, stat := range topCountryStats {
			countryName, err := regions.SharedRegionCountryDAO.FindRegionCountryName(tx, int64(stat.CountryId))
			if err != nil {
				return nil, err
			}
			result.TopCountryStats = append(result.TopCountryStats, &pb.ComposeServerStatBoardResponse_CountryStat{
				CountryName:         countryName,
				Bytes:               int64(stat.Bytes),
				CountRequests:       int64(stat.CountRequests),
				AttackBytes:         int64(stat.AttackBytes),
				CountAttackRequests: int64(stat.CountAttackRequests),
				Percent:             float32(stat.Bytes*100) / float32(totalCountryBytes),
			})
		}
	}

	// 指标
	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	var metricCategory = serverconfigs.MetricItemCategoryHTTP
	serverType, err := models.SharedServerDAO.FindEnabledServerType(tx, req.ServerId)
	if err != nil {
		return nil, err
	}
	switch serverType {
	case serverconfigs.ServerTypeTCPProxy:
		metricCategory = serverconfigs.MetricItemCategoryTCP
	case serverconfigs.ServerTypeUDPProxy:
		metricCategory = serverconfigs.MetricItemCategoryUDP
	}

	var pbCharts []*pb.MetricDataChart
	charts, ok := tasks.SharedCacheTaskManager.GetServer(tasks.CacheKeyFindServerMetricDataCharts, clusterId, req.ServerId, metricCategory)
	if ok {
		pbCharts = charts.([]*pb.MetricDataChart)
	}
	result.MetricDataCharts = pbCharts

	return result, nil
}
