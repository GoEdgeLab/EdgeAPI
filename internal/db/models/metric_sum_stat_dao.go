package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type MetricSumStatDAO dbs.DAO

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
func (this *MetricSumStatDAO) UpdateSum(tx *dbs.Tx, serverId int64, time string, itemId int64, version int32, count int64, total float32) error {
	return this.Query(tx).
		InsertOrUpdateQuickly(maps.Map{
			"serverId": serverId,
			"itemId":   itemId,
			"version":  version,
			"time":     time,
			"count":    count,
			"total":    total,
		}, maps.Map{
			"count": count,
			"total": total,
		})
}

// FindSum 查找统计数据
func (this *MetricSumStatDAO) FindSum(tx *dbs.Tx, serverId int64, time string, itemId int64, version int32) (count int64, total float32, err error) {
	one, err := this.Query(tx).
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
	return int64(one.(*MetricSumStat).Count), float32(one.(*MetricSumStat).Total), nil
}
