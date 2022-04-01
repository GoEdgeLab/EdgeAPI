// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"sync"
	"time"
)

var SharedCacheTaskManager = NewCacheTaskManager()

const (
	// global keys

	CacheKeyFindAllMetricDataCharts = "findAllMetricDataCharts"
	CacheKeyFindGlobalTopDomains    = "globalFindTopDomains"

	// cluster keys

	CacheKeyFindNodeClusterMetricDataCharts = "findNodeClusterMetricDataCharts"

	// node keys

	CacheKeyFindNodeMetricDataCharts = "findNodeMetricDataCharts"

	// server keys

	CacheKeyFindServerMetricDataCharts = "findServerMetricDataCharts"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			_ = SharedCacheTaskManager.Loop()
			SharedCacheTaskManager.Start()
		})
	})
}

type CacheTaskManager struct {
	cacheMap map[string]interface{}
	locker   sync.Mutex

	nodeTasksTooSlow bool // 记录节点任务是否太慢，如果太慢就不再后台执行

	ticker *time.Ticker
}

func NewCacheTaskManager() *CacheTaskManager {
	return &CacheTaskManager{
		cacheMap: map[string]interface{}{},
		ticker:   time.NewTicker(5 * time.Minute),
	}
}

func (this *CacheTaskManager) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			remotelogs.Error("CACHE_TASK_MANAGER", err.Error())
		}
	}
}

func (this *CacheTaskManager) Loop() error {
	var tx *dbs.Tx

	// Admin看板指标数据
	{
		value, err := this.findAllMetricDataCharts(tx)
		if err != nil {
			return err
		}
		this.locker.Lock()
		if len(value) > 0 {
			this.cacheMap[CacheKeyFindAllMetricDataCharts] = value
		}
		this.locker.Unlock()
	}

	{
		var hourFrom = timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
		var hourTo = timeutil.Format("YmdH")

		value, err := stats.SharedServerDomainHourlyStatDAO.FindTopDomainStats(tx, hourFrom, hourTo, 10)
		if err != nil {
			return err
		}

		var composedKey = CacheKeyFindGlobalTopDomains

		this.locker.Lock()
		if len(value) > 0 {
			this.cacheMap[composedKey] = value
		}
		this.locker.Unlock()
	}

	// 集群数据
	clusterIds, err := models.SharedNodeClusterDAO.FindAllEnabledNodeClusterIds(tx)
	if err != nil {
		return err
	}
	for _, clusterId := range clusterIds {
		// metric charts
		{
			var category = serverconfigs.MetricItemCategoryHTTP
			var composedKey = CacheKeyFindNodeClusterMetricDataCharts + "@" + types.String(clusterId) + "@" + category

			value, err := this.findNodeClusterMetricDataCharts(tx, clusterId, 0, 0, category)
			if err != nil {
				return err
			}
			this.locker.Lock()
			if len(value) > 0 {
				this.cacheMap[composedKey] = value
			}
			this.locker.Unlock()
		}

		// nodes
		var before = time.Now()
		if !this.nodeTasksTooSlow {
			nodeIds, err := models.SharedNodeDAO.FindEnabledNodeIdsWithClusterId(tx, clusterId)
			if err != nil {
				return err
			}
			for _, nodeId := range nodeIds {
				// metric
				{
					var category = serverconfigs.MetricItemCategoryHTTP
					var composedKey = CacheKeyFindNodeMetricDataCharts + "@" + types.String(clusterId) + "@" + types.String(nodeId) + "@" + category

					value, err := this.findNodeClusterMetricDataCharts(tx, clusterId, nodeId, 0, category)
					if err != nil {
						return err
					}
					this.locker.Lock()
					if len(value) > 0 {
						this.cacheMap[composedKey] = value
					}
					this.locker.Unlock()
				}
			}
		}
		var cost = time.Since(before).Seconds()
		if cost > 600 {
			this.nodeTasksTooSlow = true
		}
	}

	return nil
}

func (this *CacheTaskManager) Get(key string) (value interface{}, ok bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if key == CacheKeyFindAllMetricDataCharts {
		value, ok = this.cacheMap[key]
		return
	}

	return
}

func (this *CacheTaskManager) GetCluster(key string, clusterId int64, category string) (value interface{}, ok bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	var composedKey = key + "@" + types.String(clusterId) + "@" + category
	value, ok = this.cacheMap[composedKey]
	return
}

func (this *CacheTaskManager) GetNode(key string, clusterId int64, nodeId int64, category string) (value interface{}, ok bool) {
	this.locker.Lock()
	var composedKey = key + "@" + types.String(clusterId) + "@" + types.String(nodeId) + "@" + category
	value, ok = this.cacheMap[composedKey]
	this.locker.Unlock()

	if ok {
		result, err := this.findNodeClusterMetricDataCharts(nil, clusterId, nodeId, 0, category)
		if err == nil {
			value = result
		}
	}

	return
}

func (this *CacheTaskManager) GetServer(key string, clusterId int64, serverId int64, category string) (value interface{}, ok bool) {
	switch key {
	case CacheKeyFindServerMetricDataCharts:
		var tx *dbs.Tx
		v, err := this.findNodeClusterMetricDataCharts(tx, clusterId, 0, serverId, category)
		if err != nil {
			return nil, false
		}
		return v, true
	}

	return
}

func (this *CacheTaskManager) GetGlobalTopDomains() (value interface{}, ok bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	var composedKey = CacheKeyFindGlobalTopDomains
	value, ok = this.cacheMap[composedKey]
	return
}

// 所有集群的指标统计
func (this *CacheTaskManager) findAllMetricDataCharts(tx *dbs.Tx) (result []*pb.MetricDataChart, err error) {
	// 集群指标
	items, err := models.SharedMetricItemDAO.FindAllPublicItems(tx, serverconfigs.MetricItemCategoryHTTP, nil)
	if err != nil {
		return nil, err
	}
	var pbMetricCharts = []*pb.MetricDataChart{}
	for _, item := range items {
		var itemId = int64(item.Id)
		charts, err := models.SharedMetricChartDAO.FindAllEnabledCharts(tx, itemId)
		if err != nil {
			return nil, err
		}

		for _, chart := range charts {
			if !chart.IsOn {
				continue
			}

			var pbChart = &pb.MetricChart{
				Id:         int64(chart.Id),
				Name:       chart.Name,
				Type:       chart.Type,
				WidthDiv:   chart.WidthDiv,
				ParamsJSON: nil,
				IsOn:       chart.IsOn,
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
					IsOn:       item.IsOn,
				},
			}
			var pbStats = []*pb.MetricStat{}
			switch chart.Type {
			case serverconfigs.MetricChartTypeTimeLine:
				itemStats, err := models.SharedMetricStatDAO.FindLatestItemStats(tx, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				if err != nil {
					return nil, err
				}

				for _, stat := range itemStats {
					// 当前时间总和
					count, total, err := models.SharedMetricSumStatDAO.FindSumAtTime(tx, stat.Time, itemId, types.Int32(item.Version))
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
				itemStats, err := models.SharedMetricStatDAO.FindItemStatsAtLastTime(tx, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				if err != nil {
					return nil, err
				}
				for _, stat := range itemStats {
					count, total, err := models.SharedMetricSumStatDAO.FindSumAtTime(tx, stat.Time, itemId, types.Int32(item.Version))
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
			pbMetricCharts = append(pbMetricCharts, &pb.MetricDataChart{
				MetricChart: pbChart,
				MetricStats: pbStats,
			})
		}
	}
	return pbMetricCharts, nil
}

// 某个集群、节点或者服务的指标统计
func (this *CacheTaskManager) findNodeClusterMetricDataCharts(tx *dbs.Tx, clusterId int64, nodeId int64, serverId int64, category string) (result []*pb.MetricDataChart, err error) {
	// 集群指标
	clusterMetricItems, err := models.SharedNodeClusterMetricItemDAO.FindAllClusterItems(tx, clusterId, category)
	if err != nil {
		return nil, err
	}
	var pbMetricCharts = []*pb.MetricDataChart{}
	var metricItemIds = []int64{}
	var items = []*models.MetricItem{}
	for _, clusterMetricItem := range clusterMetricItems {
		if !clusterMetricItem.IsOn {
			continue
		}
		var itemId = int64(clusterMetricItem.ItemId)
		item, err := models.SharedMetricItemDAO.FindEnabledMetricItem(tx, itemId)
		if err != nil {
			return nil, err
		}
		if item == nil || !item.IsOn {
			continue
		}
		items = append(items, item)
		metricItemIds = append(metricItemIds, itemId)
	}

	publicMetricItems, err := models.SharedMetricItemDAO.FindAllPublicItems(tx, category, nil)
	if err != nil {
		return nil, err
	}
	for _, item := range publicMetricItems {
		if !item.IsOn {
			continue
		}
		if lists.ContainsInt64(metricItemIds, int64(item.Id)) {
			continue
		}
		items = append(items, item)
	}

	for _, item := range items {
		var itemId = int64(item.Id)
		charts, err := models.SharedMetricChartDAO.FindAllEnabledCharts(tx, itemId)
		if err != nil {
			return nil, err
		}

		for _, chart := range charts {
			if !chart.IsOn {
				continue
			}

			var pbChart = &pb.MetricChart{
				Id:         int64(chart.Id),
				Name:       chart.Name,
				Type:       chart.Type,
				WidthDiv:   chart.WidthDiv,
				ParamsJSON: nil,
				IsOn:       chart.IsOn,
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
					IsOn:       item.IsOn,
				},
			}
			var pbStats = []*pb.MetricStat{}
			switch chart.Type {
			case serverconfigs.MetricChartTypeTimeLine:
				var itemStats []*models.MetricStat
				if serverId > 0 {
					itemStats, err = models.SharedMetricStatDAO.FindLatestItemStatsWithServerId(tx, serverId, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				} else if nodeId > 0 {
					itemStats, err = models.SharedMetricStatDAO.FindLatestItemStatsWithNodeId(tx, nodeId, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				} else {
					itemStats, err = models.SharedMetricStatDAO.FindLatestItemStatsWithClusterId(tx, clusterId, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				}
				if err != nil {
					return nil, err
				}

				for _, stat := range itemStats {
					// 当前时间总和
					var count int64
					var total float32
					if serverId > 0 {
						count, total, err = models.SharedMetricSumStatDAO.FindServerSum(tx, serverId, stat.Time, itemId, types.Int32(item.Version))
					} else if nodeId > 0 {
						count, total, err = models.SharedMetricSumStatDAO.FindNodeSum(tx, nodeId, stat.Time, itemId, types.Int32(item.Version))
					} else {
						count, total, err = models.SharedMetricSumStatDAO.FindClusterSum(tx, clusterId, stat.Time, itemId, types.Int32(item.Version))
					}
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
				var itemStats []*models.MetricStat
				if serverId > 0 {
					itemStats, err = models.SharedMetricStatDAO.FindItemStatsWithServerIdAndLastTime(tx, serverId, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				} else if nodeId > 0 {
					itemStats, err = models.SharedMetricStatDAO.FindItemStatsWithNodeIdAndLastTime(tx, nodeId, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				} else {
					itemStats, err = models.SharedMetricStatDAO.FindItemStatsWithClusterIdAndLastTime(tx, clusterId, itemId, chart.IgnoreEmptyKeys == 1, chart.DecodeIgnoredKeys(), types.Int32(item.Version), 10)
				}
				if err != nil {
					return nil, err
				}
				for _, stat := range itemStats {
					// 当前时间总和
					var count int64
					var total float32
					if serverId > 0 {
						count, total, err = models.SharedMetricSumStatDAO.FindServerSum(tx, serverId, stat.Time, itemId, types.Int32(item.Version))
					} else if nodeId > 0 {
						count, total, err = models.SharedMetricSumStatDAO.FindNodeSum(tx, nodeId, stat.Time, itemId, types.Int32(item.Version))
					} else {
						count, total, err = models.SharedMetricSumStatDAO.FindClusterSum(tx, clusterId, stat.Time, itemId, types.Int32(item.Version))
					}
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
			pbMetricCharts = append(pbMetricCharts, &pb.MetricDataChart{
				MetricChart: pbChart,
				MetricStats: pbStats,
			})
		}
	}
	return pbMetricCharts, nil
}
