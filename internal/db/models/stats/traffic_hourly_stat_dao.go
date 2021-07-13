package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type TrafficHourlyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(24 * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedTrafficHourlyStatDAO.Clean(nil, 60) // 只保留60天
				if err != nil {
					remotelogs.Error("TrafficHourlyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

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

// IncreaseHourlyStat 增加流量
func (this *TrafficHourlyStatDAO) IncreaseHourlyStat(tx *dbs.Tx, hour string, bytes int64, cachedBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64, attackBytes int64) error {
	if len(hour) != 10 {
		return errors.New("invalid hour '" + hour + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		Param("cachedBytes", cachedBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		Param("countAttackRequests", countAttackRequests).
		Param("attackBytes", attackBytes).
		InsertOrUpdateQuickly(maps.Map{
			"hour":                hour,
			"bytes":               bytes,
			"cachedBytes":         cachedBytes,
			"countRequests":       countRequests,
			"countCachedRequests": countCachedRequests,
			"countAttackRequests": countAttackRequests,
			"attackBytes":         attackBytes,
		}, maps.Map{
			"bytes":               dbs.SQL("bytes+:bytes"),
			"cachedBytes":         dbs.SQL("cachedBytes+:cachedBytes"),
			"countRequests":       dbs.SQL("countRequests+:countRequests"),
			"countCachedRequests": dbs.SQL("countCachedRequests+:countCachedRequests"),
			"countAttackRequests": dbs.SQL("countAttackRequests+:countAttackRequests"),
			"attackBytes":         dbs.SQL("attackBytes+:attackBytes"),
		})
	if err != nil {
		return err
	}
	return nil
}

// FindHourlyStats 获取小时之间统计
func (this *TrafficHourlyStatDAO) FindHourlyStats(tx *dbs.Tx, hourFrom string, hourTo string) (result []*TrafficHourlyStat, err error) {
	ones, err := this.Query(tx).
		Between("hour", hourFrom, hourTo).
		FindAll()
	if err != nil {
		return nil, err
	}
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

// Clean 清理历史数据
func (this *TrafficHourlyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var hour = timeutil.Format("Ymd00", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("hour", hour).
		Delete()
	return err
}
