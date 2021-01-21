package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type TrafficDailyStatDAO dbs.DAO

func NewTrafficDailyStatDAO() *TrafficDailyStatDAO {
	return dbs.NewDAO(&TrafficDailyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeTrafficDailyStats",
			Model:  new(TrafficDailyStat),
			PkName: "id",
		},
	}).(*TrafficDailyStatDAO)
}

var SharedTrafficDailyStatDAO *TrafficDailyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedTrafficDailyStatDAO = NewTrafficDailyStatDAO()
	})
}

// 增加流量
func (this *TrafficDailyStatDAO) IncreaseDailyBytes(tx *dbs.Tx, day string, bytes int64) error {
	if len(day) != 8 {
		return errors.New("invalid day '" + day + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		InsertOrUpdateQuickly(maps.Map{
			"day":   day,
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
func (this *TrafficDailyStatDAO) FindDailyStats(tx *dbs.Tx, dayFrom string, dayTo string) (result []*TrafficDailyStat, err error) {
	ones, err := this.Query(tx).
		Between("day", dayFrom, dayTo).
		FindAll()
	dayMap := map[string]*TrafficDailyStat{} // day => Stat
	for _, one := range ones {
		stat := one.(*TrafficDailyStat)
		dayMap[stat.Day] = stat
	}
	days, err := utils.RangeDays(dayFrom, dayTo)
	if err != nil {
		return nil, err
	}
	for _, day := range days {
		stat, ok := dayMap[day]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &TrafficDailyStat{Day: day})
		}
	}
	return result, nil
}
