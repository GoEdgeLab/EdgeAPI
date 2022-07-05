package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type MetricStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedMetricStatDAO.Clean(nil)
				if err != nil {
					remotelogs.Error("SharedMetricStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

const MetricStatTablePartials = 20 // 表格Partial数量

func NewMetricStatDAO() *MetricStatDAO {
	return dbs.NewDAO(&MetricStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMetricStats",
			Model:  new(MetricStat),
			PkName: "id",
		},
	}).(*MetricStatDAO)
}

var SharedMetricStatDAO *MetricStatDAO

func init() {
	dbs.OnReady(func() {
		SharedMetricStatDAO = NewMetricStatDAO()
	})
}

// CreateStat 创建统计数据
func (this *MetricStatDAO) CreateStat(tx *dbs.Tx, hash string, clusterId int64, nodeId int64, serverId int64, itemId int64, keys []string, value float64, time string, version int32) error {
	hash += "@" + strconv.FormatInt(nodeId, 10)
	var keysString string
	if len(keys) > 0 {
		keysJSON, err := json.Marshal(keys)
		if err != nil {
			return err
		}
		keysString = string(keysJSON)
	} else {
		keysString = "[]"
	}
	err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Param("value", value).
		InsertOrUpdateQuickly(maps.Map{
			"hash":       hash,
			"clusterId":  clusterId,
			"nodeId":     nodeId,
			"serverId":   serverId,
			"itemId":     itemId,
			"value":      value,
			"time":       time,
			"version":    version,
			"keys":       keysString,
			"createdDay": timeutil.Format("Ymd"),
		}, maps.Map{
			"value": value,
		})
	if err != nil {
		if this.canIgnore(err) {
			return nil
		}
		return err
	}

	return SharedMetricItemDAO.UpdateMetricLastTime(tx, itemId, time)
}

// DeleteOldVersionItemStats 删除以前版本的统计数据
func (this *MetricStatDAO) DeleteOldVersionItemStats(tx *dbs.Tx, itemId int64, version int32) error {
	return this.runBatch(func(table string, locker *sync.Mutex) error {
		_, err := this.Query(tx).
			Table(table).
			Attr("itemId", itemId).
			Where("version<:version").
			Param("version", version).
			Delete()
		return err
	})
}

// DeleteItemStats 删除某个指标相关的统计数据
func (this *MetricStatDAO) DeleteItemStats(tx *dbs.Tx, itemId int64) error {
	return this.runBatch(func(table string, locker *sync.Mutex) error {
		_, err := this.Query(tx).
			Table(table).
			Attr("itemId", itemId).
			Delete()
		return err
	})
}

// DeleteNodeItemStats 删除某个节点的统计数据
func (this *MetricStatDAO) DeleteNodeItemStats(tx *dbs.Tx, nodeId int64, serverId int64, itemId int64, time string) error {
	if serverId > 0 {
		_, err := this.Query(tx).
			Table(this.partialTable(serverId)).
			Attr("nodeId", nodeId).
			Attr("serverId", serverId).
			Attr("itemId", itemId).
			Attr("time", time).
			Delete()
		if this.canIgnore(err) {
			return nil
		}
		return err
	}

	err := this.runBatch(func(table string, locker *sync.Mutex) error {
		_, err := this.Query(tx).
			Table(table).
			Attr("nodeId", nodeId).
			Attr("serverId", serverId).
			Attr("itemId", itemId).
			Attr("time", time).
			Delete()
		return err
	})

	return err
}

// CountItemStats 计算统计数据数量
func (this *MetricStatDAO) CountItemStats(tx *dbs.Tx, itemId int64, version int32) (int64, error) {
	var total int64 = 0

	err := this.runBatch(func(table string, locker *sync.Mutex) error {
		count, err := this.Query(tx).
			Table(table).
			Attr("itemId", itemId).
			Attr("version", version).
			Count()
		if err != nil {
			return err
		}
		atomic.AddInt64(&total, count)
		return nil
	})
	if err != nil {
		return 0, err
	}

	return total, nil
}

// ListItemStats 列出单页统计数据
func (this *MetricStatDAO) ListItemStats(tx *dbs.Tx, itemId int64, version int32, offset int64, size int64) (result []*MetricStat, err error) {
	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		var partialResult = []*MetricStat{}
		_, err = this.Query(tx).
			Table(table).
			Attr("itemId", itemId).
			Attr("version", version).
			Offset(offset).
			Limit(size).
			Desc("time").
			Desc("serverId").
			Desc("value").
			Slice(&partialResult).
			FindAll()
		if err != nil {
			return err
		}
		locker.Lock()
		result = append(result, partialResult...)
		locker.Unlock()
		return nil
	})
	if err != nil {
		return nil, err
	}

	result = this.mergeStats(result)

	sort.Slice(result, func(i, j int) bool {
		if result[i].Time > result[j].Time {
			return true
		}
		if result[i].ServerId > result[j].ServerId {
			return true
		}
		if result[i].Value > result[j].Value {
			return true
		}
		return false
	})

	return
}

// FindItemStatsAtLastTime 取得所有集群最近一次计时前 N 个数据
// 适合每条数据中包含不同的Key的场景
func (this *MetricStatDAO) FindItemStatsAtLastTime(tx *dbs.Tx, itemId int64, ignoreEmptyKeys bool, ignoreKeys []string, version int32, size int64) (result []*MetricStat, err error) {
	// 最近一次时间
	lastTime, err := SharedMetricItemDAO.FindMetricLastTime(tx, itemId)
	if err != nil || len(lastTime) == 0 {
		return nil, err
	}

	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		var partialResult = []*MetricStat{}
		var query = this.Query(tx).
			Table(table).
			Attr("itemId", itemId).
			Attr("version", version).
			Attr("time", lastTime).
			// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
			// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
			Result("MIN(time) AS time", "SUM(value) AS value", "keys").
			Desc("value").
			Group("keys").
			Limit(size).
			Slice(&partialResult)
		if ignoreEmptyKeys {
			query.Where("NOT JSON_CONTAINS(`keys`, '\"\"')")
		}
		if len(ignoreKeys) > 0 {
			ignoreKeysJSON, err := json.Marshal(ignoreKeys)
			if err != nil {
				return err
			}
			query.Where("NOT JSON_CONTAINS(:ignoredKeys, JSON_EXTRACT(`keys`, '$[0]'))") // TODO $[0] 需要换成keys中的primary key位置
			query.Param("ignoredKeys", string(ignoreKeysJSON))
		}
		_, err = query.
			FindAll()
		if err != nil {
			return err
		}

		locker.Lock()
		result = append(result, partialResult...)
		locker.Unlock()

		return nil
	})
	if err != nil {
		return nil, err
	}

	result = this.mergeStats(result)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	return
}

// FindItemStatsWithClusterIdAndLastTime 取得集群最近一次计时前 N 个数据
// 适合每条数据中包含不同的Key的场景
func (this *MetricStatDAO) FindItemStatsWithClusterIdAndLastTime(tx *dbs.Tx, clusterId int64, itemId int64, ignoreEmptyKeys bool, ignoreKeys []string, version int32, size int64) (result []*MetricStat, err error) {
	lastTime, err := SharedMetricItemDAO.FindMetricLastTime(tx, itemId)
	if err != nil || len(lastTime) == 0 {
		return nil, err
	}

	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		var partialResult = []*MetricStat{}
		var query = this.Query(tx).
			Table(table).
			UseIndex("cluster_item_time").
			Attr("clusterId", clusterId).
			Attr("itemId", itemId).
			Attr("version", version).
			Attr("time", lastTime).
			// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
			// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
			Result("MIN(time) AS time", "SUM(value) AS value", "keys").
			Desc("value").
			Group("keys").
			Limit(size).
			Slice(&partialResult)
		if ignoreEmptyKeys {
			query.Where("NOT JSON_CONTAINS(`keys`, '\"\"')")
		}
		if len(ignoreKeys) > 0 {
			ignoreKeysJSON, err := json.Marshal(ignoreKeys)
			if err != nil {
				return err
			}
			query.Where("NOT JSON_CONTAINS(:ignoredKeys, JSON_EXTRACT(`keys`, '$[0]'))") // TODO $[0] 需要换成keys中的primary key位置
			query.Param("ignoredKeys", string(ignoreKeysJSON))
		}

		_, err = query.
			FindAll()
		if err != nil {
			return err
		}

		locker.Lock()
		result = append(result, partialResult...)
		locker.Unlock()

		return nil
	})
	if err != nil {
		return nil, err
	}

	result = this.mergeStats(result)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	return
}

// FindItemStatsWithNodeIdAndLastTime 取得节点最近一次计时前 N 个数据
// 适合每条数据中包含不同的Key的场景
func (this *MetricStatDAO) FindItemStatsWithNodeIdAndLastTime(tx *dbs.Tx, nodeId int64, itemId int64, ignoreEmptyKeys bool, ignoreKeys []string, version int32, size int64) (result []*MetricStat, err error) {
	// 最近一次时间
	lastTime, err := SharedMetricItemDAO.FindMetricLastTime(tx, itemId)
	if err != nil {
		return nil, err
	}

	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		var partialResult = []*MetricStat{}
		var query = this.Query(tx).
			Table(table).
			UseIndex("node_item_time").
			Attr("nodeId", nodeId).
			Attr("itemId", itemId).
			Attr("version", version).
			Attr("time", lastTime).
			// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
			// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
			Result("MIN(time) AS time", "SUM(value) AS value", "keys").
			Desc("value").
			Group("keys").
			Limit(size).
			Slice(&partialResult)
		if ignoreEmptyKeys {
			query.Where("NOT JSON_CONTAINS(`keys`, '\"\"')")
		}
		if len(ignoreKeys) > 0 {
			ignoreKeysJSON, err := json.Marshal(ignoreKeys)
			if err != nil {
				return err
			}
			query.Where("NOT JSON_CONTAINS(:ignoredKeys, JSON_EXTRACT(`keys`, '$[0]'))") // TODO $[0] 需要换成keys中的primary key位置
			query.Param("ignoredKeys", string(ignoreKeysJSON))
		}

		_, err = query.
			FindAll()
		if err != nil {
			return err
		}

		locker.Lock()
		result = append(result, partialResult...)
		locker.Unlock()

		return nil
	})

	if err != nil {
		return nil, err
	}

	result = this.mergeStats(result)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	return
}

// FindItemStatsWithServerIdAndLastTime 取得服务最近一次计时前 N 个数据
// 适合每条数据中包含不同的Key的场景
func (this *MetricStatDAO) FindItemStatsWithServerIdAndLastTime(tx *dbs.Tx, serverId int64, itemId int64, ignoreEmptyKeys bool, ignoreKeys []string, version int32, size int64) (result []*MetricStat, err error) {
	// 最近一次时间
	lastTime, err := SharedMetricItemDAO.FindMetricLastTime(tx, itemId)
	if err != nil || len(lastTime) == 0 {
		return nil, err
	}

	var query = this.Query(tx).
		Table(this.partialTable(serverId)).
		UseIndex("server_item_time").
		Attr("serverId", serverId).
		Attr("itemId", itemId).
		Attr("version", version).
		Attr("time", lastTime).
		// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
		// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
		Result("MIN(time) AS time", "SUM(value) AS value", "keys").
		Desc("value").
		Group("keys").
		Limit(size).
		Slice(&result)
	if ignoreEmptyKeys {
		query.Where("NOT JSON_CONTAINS(`keys`, '\"\"')")
	}
	if len(ignoreKeys) > 0 {
		ignoreKeysJSON, err := json.Marshal(ignoreKeys)
		if err != nil {
			return nil, err
		}
		query.Where("NOT JSON_CONTAINS(:ignoredKeys, JSON_EXTRACT(`keys`, '$[0]'))") // TODO $[0] 需要换成keys中的primary key位置
		query.Param("ignoredKeys", string(ignoreKeysJSON))
	}

	_, err = query.
		FindAll()
	return
}

// FindLatestItemStats 取得所有集群上最近 N 个时间的数据
// 适合同个Key在不同时间段的变化场景
func (this *MetricStatDAO) FindLatestItemStats(tx *dbs.Tx, itemId int64, ignoreEmptyKeys bool, ignoreKeys []string, version int32, size int64) (result []*MetricStat, err error) {
	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		var partialResult = []*MetricStat{}
		var query = this.Query(tx).
			Table(table).
			Attr("itemId", itemId).
			Attr("version", version).
			// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
			// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
			Result("time", "SUM(value) AS value", "MIN(`keys`) AS `keys`").
			Desc("time").
			Group("time").
			Limit(size).
			Slice(&partialResult)
		if ignoreEmptyKeys {
			query.Where("NOT JSON_CONTAINS(`keys`, '\"\"')")
		}
		if len(ignoreKeys) > 0 {
			ignoreKeysJSON, err := json.Marshal(ignoreKeys)
			if err != nil {
				return err
			}
			query.Where("NOT JSON_CONTAINS(:ignoredKeys, JSON_EXTRACT(`keys`, '$[0]'))") // TODO $[0] 需要换成keys中的primary key位置
			query.Param("ignoredKeys", string(ignoreKeysJSON))
		}

		_, err = query.
			FindAll()
		if err != nil {
			return err
		}

		locker.Lock()
		result = append(result, partialResult...)
		locker.Unlock()

		return nil
	})

	if err != nil {
		return nil, err
	}

	result = this.mergeStats(result)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Time > result[j].Time
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	lists.Reverse(result)

	return
}

// FindLatestItemStatsWithClusterId 取得集群最近 N 个时间的数据
// 适合同个Key在不同时间段的变化场景
func (this *MetricStatDAO) FindLatestItemStatsWithClusterId(tx *dbs.Tx, clusterId int64, itemId int64, ignoreEmptyKeys bool, ignoreKeys []string, version int32, size int64) (result []*MetricStat, err error) {
	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		var partialResult = []*MetricStat{}
		var query = this.Query(tx).
			Table(table).
			Attr("clusterId", clusterId).
			Attr("itemId", itemId).
			Attr("version", version).
			// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
			// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
			Result("time", "SUM(value) AS value", "MIN(`keys`) AS `keys`").
			Desc("time").
			Group("time").
			Limit(size).
			Slice(&partialResult)
		if ignoreEmptyKeys {
			query.Where("NOT JSON_CONTAINS(`keys`, '\"\"')")
		}
		if len(ignoreKeys) > 0 {
			ignoreKeysJSON, err := json.Marshal(ignoreKeys)
			if err != nil {
				return err
			}
			query.Where("NOT JSON_CONTAINS(:ignoredKeys, JSON_EXTRACT(`keys`, '$[0]'))") // TODO $[0] 需要换成keys中的primary key位置
			query.Param("ignoredKeys", string(ignoreKeysJSON))
		}

		_, err = query.
			FindAll()
		if err != nil {
			return err
		}

		locker.Lock()
		result = append(result, partialResult...)
		locker.Unlock()

		return nil
	})

	if err != nil {
		return nil, err
	}

	result = this.mergeStats(result)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Time > result[j].Time
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	lists.Reverse(result)
	return
}

// FindLatestItemStatsWithNodeId 取得节点最近 N 个时间的数据
// 适合同个Key在不同时间段的变化场景
func (this *MetricStatDAO) FindLatestItemStatsWithNodeId(tx *dbs.Tx, nodeId int64, itemId int64, ignoreEmptyKeys bool, ignoreKeys []string, version int32, size int64) (result []*MetricStat, err error) {
	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		var partialResult = []*MetricStat{}
		var query = this.Query(tx).
			Table(table).
			Attr("nodeId", nodeId).
			Attr("itemId", itemId).
			Attr("version", version).
			// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
			// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
			Result("time", "SUM(value) AS value", "MIN(`keys`) AS `keys`").
			Desc("time").
			Group("time").
			Limit(size).
			Slice(&partialResult)
		if ignoreEmptyKeys {
			query.Where("NOT JSON_CONTAINS(`keys`, '\"\"')")
		}
		if len(ignoreKeys) > 0 {
			ignoreKeysJSON, err := json.Marshal(ignoreKeys)
			if err != nil {
				return err
			}
			query.Where("NOT JSON_CONTAINS(:ignoredKeys, JSON_EXTRACT(`keys`, '$[0]'))") // TODO $[0] 需要换成keys中的primary key位置
			query.Param("ignoredKeys", string(ignoreKeysJSON))
		}

		_, err = query.
			FindAll()
		if err != nil {
			return err
		}

		locker.Lock()
		result = append(result, partialResult...)
		locker.Unlock()

		return nil
	})

	result = this.mergeStats(result)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Time > result[j].Time
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	lists.Reverse(result)
	return
}

// FindLatestItemStatsWithServerId 取得服务最近 N 个时间的数据
// 适合同个Key在不同时间段的变化场景
func (this *MetricStatDAO) FindLatestItemStatsWithServerId(tx *dbs.Tx, serverId int64, itemId int64, ignoreEmptyKeys bool, ignoreKeys []string, version int32, size int64) (result []*MetricStat, err error) {
	var query = this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Attr("itemId", itemId).
		Attr("version", version).
		// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
		// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
		Result("time", "SUM(value) AS value", "MIN(`keys`) AS `keys`").
		Desc("time").
		Group("time").
		Limit(size).
		Slice(&result)
	if ignoreEmptyKeys {
		query.Where("NOT JSON_CONTAINS(`keys`, '\"\"')")
	}
	if len(ignoreKeys) > 0 {
		ignoreKeysJSON, err := json.Marshal(ignoreKeys)
		if err != nil {
			return nil, err
		}
		query.Where("NOT JSON_CONTAINS(:ignoredKeys, JSON_EXTRACT(`keys`, '$[0]'))") // TODO $[0] 需要换成keys中的primary key位置
		query.Param("ignoredKeys", string(ignoreKeysJSON))
	}

	_, err = query.
		FindAll()
	if err != nil {
		return nil, err
	}
	lists.Reverse(result)
	return
}

// Clean 清理数据
func (this *MetricStatDAO) Clean(tx *dbs.Tx) error {
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
				err := this.runBatch(func(table string, locker *sync.Mutex) error {
					_, err := this.Query(tx).
						Table(table).
						Attr("itemId", item.Id).
						Lte("createdDay", expiresDay).
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
func (this *MetricStatDAO) partialTable(serverId int64) string {
	return this.Table + "_" + types.String(serverId%int64(MetricStatTablePartials))
}

// 批量执行
func (this *MetricStatDAO) runBatch(f func(table string, locker *sync.Mutex) error) error {
	var locker = &sync.Mutex{}
	var wg = sync.WaitGroup{}
	wg.Add(MetricStatTablePartials)
	var resultErr error
	for i := 0; i < MetricStatTablePartials; i++ {
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

// 合并统计数据
func (this *MetricStatDAO) mergeStats(stats []*MetricStat) (result []*MetricStat) {
	var m = map[string]*MetricStat{} // key+time => *MetricStat
	for _, stat := range stats {
		var uniqueKey = stat.Time + "@" + string(stat.Keys)
		oldStat, ok := m[uniqueKey]
		if !ok {
			result = append(result, stat)
			m[uniqueKey] = stat
		} else {
			oldStat.Value += stat.Value
		}
	}
	return result
}

// 检查错误是否可以忽略
func (this *MetricStatDAO) canIgnore(err error) bool {
	if err == nil {
		return true
	}

	// 忽略 Error 1213: Deadlock found 错误
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok && mysqlErr.Number == 1213 {
		return true
	}

	return false
}
