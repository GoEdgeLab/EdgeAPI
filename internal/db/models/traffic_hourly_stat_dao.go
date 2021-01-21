package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type TrafficHourlyStatDAO dbs.DAO

func NewTrafficHourlyStatDAO() *TrafficHourlyStatDAO {
	return dbs.NewDAO(&TrafficHourlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeTrafficHourlyStats",
			Model:  new(TrafficHourlyStat),
			PkName: "id",
		},
	}).(*TrafficHourlyStatDAO)
}

var SharedTrafficHourlyStatDAO *TrafficHourlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedTrafficHourlyStatDAO = NewTrafficHourlyStatDAO()
	})
}

// 增加流量
func (this *TrafficHourlyStatDAO) IncreaseHourlyBytes(tx *dbs.Tx, hour string, bytes int64) error {
	if len(hour) != 10 {
		return errors.New("invalid hour '" + hour + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		InsertOrUpdateQuickly(maps.Map{
			"hour":  hour,
			"bytes": bytes,
		}, maps.Map{
			"bytes": dbs.SQL("bytes+:bytes"),
		})
	if err != nil {
		return err
	}
	return nil
}

// 获取日期之间统计
func (this *TrafficHourlyStatDAO) FindHourlyStats(tx *dbs.Tx, hourFrom string, hourTo string) (result []*TrafficHourlyStat, err error) {
	ones, err := this.Query(tx).
		Between("hour", hourFrom, hourTo).
		FindAll()
	hourMap := map[string]*TrafficHourlyStat{} // hour => Stat
	for _, one := range ones {
		stat := one.(*TrafficHourlyStat)
		hourMap[stat.Hour] = stat
	}
	hours, err := utils.RangeHours(hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, hour := range hours {
		stat, ok := hourMap[hour]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &TrafficHourlyStat{Hour: hour})
		}
	}
	return result, nil
}
