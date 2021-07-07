package models

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
	"strconv"
)

type MetricStatDAO dbs.DAO

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
	return this.Query(tx).
		Param("value", value).
		InsertOrUpdateQuickly(maps.Map{
			"hash":      hash,
			"clusterId": clusterId,
			"nodeId":    nodeId,
			"serverId":  serverId,
			"itemId":    itemId,
			"value":     value,
			"time":      time,
			"version":   version,
			"keys":      keysString,
		}, maps.Map{
			"value": value,
		})
}

// DeleteOldItemStats 删除以前版本的统计数据
func (this *MetricStatDAO) DeleteOldItemStats(tx *dbs.Tx, itemId int64, version int32) error {
	_, err := this.Query(tx).
		Attr("itemId", itemId).
		Where("version<:version").
		Delete()
	return err
}

// DeleteItemStats 删除某个指标相关的统计数据
func (this *MetricStatDAO) DeleteItemStats(tx *dbs.Tx, itemId int64) error {
	_, err := this.Query(tx).
		Attr("itemId", itemId).
		Delete()
	return err
}

// CountItemStats 计算统计数据数量
func (this *MetricStatDAO) CountItemStats(tx *dbs.Tx, itemId int64, version int32) (int64, error) {
	return this.Query(tx).
		Attr("itemId", itemId).
		Attr("version", version).
		Count()
}

// ListItemStats 列出单页统计数据
func (this *MetricStatDAO) ListItemStats(tx *dbs.Tx, itemId int64, version int32, offset int64, size int64) (result []*MetricStat, err error) {
	_, err = this.Query(tx).
		Attr("itemId", itemId).
		Attr("version", version).
		Offset(offset).
		Limit(size).
		Desc("time").
		Desc("serverId").
		Desc("value").
		Slice(&result).
		FindAll()
	return
}

// FindItemStatsWithClusterIdAndLastTime 取得集群最近一次计时前 N 个数据
// 适合每条数据中包含不同的Key的场景
func (this *MetricStatDAO) FindItemStatsWithClusterIdAndLastTime(tx *dbs.Tx, clusterId int64, itemId int64, version int32, size int64) (result []*MetricStat, err error) {
	// 最近一次时间
	statOne, err := this.Query(tx).
		Attr("itemId", itemId).
		Attr("version", version).
		DescPk().
		Find()
	if err != nil {
		return nil, err
	}
	if statOne == nil {
		return nil, nil
	}
	var lastStat = statOne.(*MetricStat)
	var lastTime = lastStat.Time

	_, err = this.Query(tx).
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
		Slice(&result).
		FindAll()
	return
}

// FindItemStatsWithNodeIdAndLastTime 取得节点最近一次计时前 N 个数据
// 适合每条数据中包含不同的Key的场景
func (this *MetricStatDAO) FindItemStatsWithNodeIdAndLastTime(tx *dbs.Tx, nodeId int64, itemId int64, version int32, size int64) (result []*MetricStat, err error) {
	// 最近一次时间
	statOne, err := this.Query(tx).
		Attr("itemId", itemId).
		Attr("version", version).
		DescPk().
		Find()
	if err != nil {
		return nil, err
	}
	if statOne == nil {
		return nil, nil
	}
	var lastStat = statOne.(*MetricStat)
	var lastTime = lastStat.Time

	_, err = this.Query(tx).
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
		Slice(&result).
		FindAll()
	return
}

// FindItemStatsWithServerIdAndLastTime 取得节点最近一次计时前 N 个数据
// 适合每条数据中包含不同的Key的场景
func (this *MetricStatDAO) FindItemStatsWithServerIdAndLastTime(tx *dbs.Tx, serverId int64, itemId int64, version int32, size int64) (result []*MetricStat, err error) {
	// 最近一次时间
	statOne, err := this.Query(tx).
		Attr("itemId", itemId).
		Attr("version", version).
		DescPk().
		Find()
	if err != nil {
		return nil, err
	}
	if statOne == nil {
		return nil, nil
	}
	var lastStat = statOne.(*MetricStat)
	var lastTime = lastStat.Time

	_, err = this.Query(tx).
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
		Slice(&result).
		FindAll()
	return
}

// FindLatestItemStatsWithClusterId 取得集群最近 N 个时间的数据
// 适合同个Key在不同时间段的变化场景
func (this *MetricStatDAO) FindLatestItemStatsWithClusterId(tx *dbs.Tx, clusterId int64, itemId int64, version int32, size int64) (result []*MetricStat, err error) {
	_, err = this.Query(tx).
		Attr("clusterId", clusterId).
		Attr("itemId", itemId).
		Attr("version", version).
		// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
		// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
		Result("time", "SUM(value) AS value", "MIN(`keys`) AS `keys`").
		Desc("time").
		Group("time").
		Limit(size).
		Slice(&result).
		FindAll()
	if err != nil {
		return nil, err
	}
	lists.Reverse(result)
	return
}

// FindLatestItemStatsWithNodeId 取得节点最近 N 个时间的数据
// 适合同个Key在不同时间段的变化场景
func (this *MetricStatDAO) FindLatestItemStatsWithNodeId(tx *dbs.Tx, nodeId int64, itemId int64, version int32, size int64) (result []*MetricStat, err error) {
	_, err = this.Query(tx).
		Attr("nodeId", nodeId).
		Attr("itemId", itemId).
		Attr("version", version).
		// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
		// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
		Result("time", "SUM(value) AS value", "MIN(`keys`) AS `keys`").
		Desc("time").
		Group("time").
		Limit(size).
		Slice(&result).
		FindAll()
	if err != nil {
		return nil, err
	}
	lists.Reverse(result)
	return
}

// FindLatestItemStatsWithServerId 取得服务最近 N 个时间的数据
// 适合同个Key在不同时间段的变化场景
func (this *MetricStatDAO) FindLatestItemStatsWithServerId(tx *dbs.Tx, serverId int64, itemId int64, version int32, size int64) (result []*MetricStat, err error) {
	_, err = this.Query(tx).
		Attr("serverId", serverId).
		Attr("itemId", itemId).
		Attr("version", version).
		// TODO 增加更多聚合算法，比如 AVG、MEDIAN、MIN、MAX 等
		// TODO 这里的 MIN(`keys`) 在MySQL8中可以换成FIRST_VALUE
		Result("time", "SUM(value) AS value", "MIN(`keys`) AS `keys`").
		Desc("time").
		Group("time").
		Limit(size).
		Slice(&result).
		FindAll()
	if err != nil {
		return nil, err
	}
	lists.Reverse(result)
	return
}
