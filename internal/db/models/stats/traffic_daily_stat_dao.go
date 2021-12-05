package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type TrafficDailyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedTrafficDailyStatDAO.Clean(nil, 30) // 只保留N天
				if err != nil {
					remotelogs.Error("TrafficDailyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

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

// IncreaseDailyStat 增加统计数据
func (this *TrafficDailyStatDAO) IncreaseDailyStat(tx *dbs.Tx, day string, bytes int64, cachedBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64, attackBytes int64) error {
	if len(day) != 8 {
		return errors.New("invalid day '" + day + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		Param("cachedBytes", cachedBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		Param("countAttackRequests", countAttackRequests).
		Param("attackBytes", attackBytes).
		InsertOrUpdateQuickly(maps.Map{
			"day":                 day,
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

// FindDailyStats 获取日期之间统计
func (this *TrafficDailyStatDAO) FindDailyStats(tx *dbs.Tx, dayFrom string, dayTo string) (result []*TrafficDailyStat, err error) {
	ones, err := this.Query(tx).
		Between("day", dayFrom, dayTo).
		FindAll()
	if err != nil {
		return nil, err
	}
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

// Clean 清理历史数据
func (this *TrafficDailyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("day", day).
		Delete()
	return err
}
