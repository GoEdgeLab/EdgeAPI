package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"sync"
	"time"
)

type MetricSumStatDAO dbs.DAO

const MetricSumStatTablePartials = 20 // 表格Partial数量

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedMetricSumStatDAO.Clean(nil)
				if err != nil {
					remotelogs.Error("SharedMetricSumStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

func NewMetricSumStatDAO() *MetricSumStatDAO {
	return dbs.NewDAO(&MetricSumStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMetricSumStats",
			Model:  new(MetricSumStat),
			PkName: "id",
		},
	}).(*MetricSumStatDAO)
}

var SharedMetricSumStatDAO *MetricSumStatDAO

func init() {
	dbs.OnReady(func() {
		SharedMetricSumStatDAO = NewMetricSumStatDAO()
	})
}

// UpdateSum 更新统计数据
func (this *MetricSumStatDAO) UpdateSum(tx *dbs.Tx, clusterId int64, nodeId int64, serverId int64, time string, itemId int64, version int32, count int64, total float32) error {
	err := this.Query(tx).
		Table(this.partialTable(serverId)).
		InsertOrUpdateQuickly(maps.Map{
			"clusterId":  clusterId,
			"nodeId":     nodeId,
			"serverId":   serverId,
			"itemId":     itemId,
			"version":    version,
			"time":       time,
			"count":      count,
			"total":      total,
			"createdDay": timeutil.Format("Ymd"),
		}, maps.Map{
			"count": count,
			"total": total,
		})
	if this.canIgnore(err) {
		return nil
	}
	return err
}

// FindNodeServerSum 查找某个服务在某个节点上的统计数据
func (this *MetricSumStatDAO) FindNodeServerSum(tx *dbs.Tx, nodeId int64, serverId int64, time string, itemId int64, version int32) (count int64, total float32, err error) {
	one, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("nodeId", nodeId).
		Attr("serverId", serverId).
		Attr("time", time).
		Attr("itemId", itemId).
		Attr("version", version).
		Find()
	if err != nil {
		return 0, 0, err
	}
	if one == nil {
		return
	}

	count = int64(one.(*MetricSumStat).Count)
	total = float32(one.(*MetricSumStat).Total)

	return
}

// FindSumAtTime 查找某个时间的统计数据
func (this *MetricSumStatDAO) FindSumAtTime(tx *dbs.Tx, time string, itemId int64, version int32) (count int64, total float32, err error) {
	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		one, err := this.Query(tx).
			Table(table).
			Attr("time", time).
			Attr("itemId", itemId).
			Attr("version", version).
			Result("SUM(count) AS `count`, SUM(total) AS total").
			Find()
		if err != nil {
			return err
		}
		if one == nil {
			return nil
		}
		locker.Lock()
		count += int64(one.(*MetricSumStat).Count)
		total += float32(one.(*MetricSumStat).Total)
		locker.Unlock()
		return nil
	})
	return
}

// FindServerSum 查找某个服务的统计数据
func (this *MetricSumStatDAO) FindServerSum(tx *dbs.Tx, serverId int64, time string, itemId int64, version int32) (count int64, total float32, err error) {
	one, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		UseIndex("server_item_time").
		Attr("serverId", serverId).
		Attr("time", time).
		Attr("itemId", itemId).
		Attr("version", version).
		Result("SUM(count) AS `count`, SUM(total) AS total").
		Find()
	if err != nil {
		return 0, 0, err
	}
	if one == nil {
		return
	}
	return int64(one.(*MetricSumStat).Count), float32(one.(*MetricSumStat).Total), nil
}

// FindClusterSum 查找集群上的统计数据
func (this *MetricSumStatDAO) FindClusterSum(tx *dbs.Tx, clusterId int64, time string, itemId int64, version int32) (count int64, total float32, err error) {
	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		one, err := this.Query(tx).
			Table(table).
			UseIndex("cluster_item_time").
			Attr("clusterId", clusterId).
			Attr("time", time).
			Attr("itemId", itemId).
			Attr("version", version).
			Result("SUM(count) AS `count`, SUM(total) AS total").
			Find()
		if err != nil {
			return err
		}
		if one == nil {
			return nil
		}
		locker.Lock()
		count += int64(one.(*MetricSumStat).Count)
		total += float32(one.(*MetricSumStat).Total)
		locker.Unlock()
		return nil
	})

	return
}

// FindNodeSum 查找节点上的统计数据
func (this *MetricSumStatDAO) FindNodeSum(tx *dbs.Tx, nodeId int64, time string, itemId int64, version int32) (count int64, total float32, err error) {
	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		one, err := this.Query(tx).
			Table(table).
			UseIndex("node_item_time").
			Attr("nodeId", nodeId).
			Attr("time", time).
			Attr("itemId", itemId).
			Attr("version", version).
			Result("SUM(count) AS `count`, SUM(total) AS total").
			Find()
		if err != nil {
			return err
		}
		if one == nil {
			return nil
		}
		locker.Lock()
		count += int64(one.(*MetricSumStat).Count)
		total += float32(one.(*MetricSumStat).Total)
		locker.Unlock()
		return nil
	})

	return
}

// DeleteItemStats 删除某个指标相关的统计数据
func (this *MetricSumStatDAO) DeleteItemStats(tx *dbs.Tx, itemId int64) error {
	return this.runBatch(func(table string, locker *sync.Mutex) error {
		_, err := this.Query(tx).
			Table(table).
			Attr("itemId", itemId).
			Delete()
		return err
	})
}

// Clean 清理数据
func (this *MetricSumStatDAO) Clean(tx *dbs.Tx) error {
	for _, category := range serverconfigs.FindAllMetricItemCategoryCodes() {
		var offset int64 = 0
		var size int64 = 100
		for {
			items, err := SharedMetricItemDAO.ListEnabledItems(tx, category, offset, size)
			if err != nil {
				return err
			}
			for _, item := range items {
				var config = &serverconfigs.MetricItemConfig{
					Id:            int64(item.Id),
					Period:        int(item.Period),
					PeriodUnit:    item.PeriodUnit,
					ExpiresPeriod: int(item.ExpiresPeriod),
				}
				var expiresDay = config.ServerExpiresDay()
				err = this.runBatch(func(table string, locker *sync.Mutex) error {
					_, err := this.Query(tx).
						Table(table).
						Attr("itemId", item.Id).
						Where("(createdDay IS NULL OR createdDay<:day)").
						Param("day", expiresDay).
						Limit(10_000). // 一次性不要删除太多，防止阻塞其他操作
						Delete()
					return err
				})
				if err != nil {
					return err
				}
			}

			if len(items) == 0 {
				break
			}

			offset += size
		}
	}
	return nil
}

// 获取分区表
func (this *MetricSumStatDAO) partialTable(serverId int64) string {
	return this.Table + "_" + types.String(serverId%int64(MetricSumStatTablePartials))
}

// 批量执行
func (this *MetricSumStatDAO) runBatch(f func(table string, locker *sync.Mutex) error) error {
	var locker = &sync.Mutex{}
	var wg = sync.WaitGroup{}
	wg.Add(MetricSumStatTablePartials)
	var resultErr error
	for i := 0; i < MetricSumStatTablePartials; i++ {
		var table = this.partialTable(int64(i))
		go func(table string) {
			defer wg.Done()

			err := f(table, locker)
			if err != nil {
				resultErr = err
			}
		}(table)
	}
	wg.Wait()
	return resultErr
}

// 检查错误是否可以忽略
func (this *MetricSumStatDAO) canIgnore(err error) bool {
	if err == nil {
		return true
	}

	// 忽略 Error 1213: Deadlock found 错误
	return CheckSQLErrCode(err, 1213)
}
