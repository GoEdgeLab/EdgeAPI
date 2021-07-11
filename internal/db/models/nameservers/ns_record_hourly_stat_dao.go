package nameservers

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

type NSRecordHourlyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(24 * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedNSRecordHourlyStatDAO.Clean(nil, 60) // 只保留60天
				if err != nil {
					remotelogs.Error("NodeClusterTrafficDailyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

func NewNSRecordHourlyStatDAO() *NSRecordHourlyStatDAO {
	return dbs.NewDAO(&NSRecordHourlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSRecordHourlyStats",
			Model:  new(NSRecordHourlyStat),
			PkName: "id",
		},
	}).(*NSRecordHourlyStatDAO)
}

var SharedNSRecordHourlyStatDAO *NSRecordHourlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedNSRecordHourlyStatDAO = NewNSRecordHourlyStatDAO()
	})
}

// IncreaseHourlyStat 增加统计数据
func (this *NSRecordHourlyStatDAO) IncreaseHourlyStat(tx *dbs.Tx, clusterId int64, nodeId int64, hour string, domainId int64, recordId int64, countRequests int64, bytes int64) error {
	if len(hour) != 10 {
		return errors.New("invalid hour '" + hour + "'")
	}
	return this.Query(tx).
		Param("countRequests", countRequests).
		Param("bytes", bytes).
		InsertOrUpdateQuickly(maps.Map{
			"clusterId":     clusterId,
			"nodeId":        nodeId,
			"domainId":      domainId,
			"recordId":      recordId,
			"day":           hour[:8],
			"hour":          hour,
			"countRequests": countRequests,
			"bytes":         bytes,
		}, maps.Map{
			"countRequests": dbs.SQL("countRequests+:countRequests"),
			"bytes":         dbs.SQL("bytes+:bytes"),
		})
}

// FindHourlyStats 按小时统计
func (this *NSRecordHourlyStatDAO) FindHourlyStats(tx *dbs.Tx, hourFrom string, hourTo string) (result []*NSRecordHourlyStat, err error) {
	ones, err := this.Query(tx).
		Result("hour", "SUM(countRequests) AS countRequests", "SUM(bytes) AS bytes").
		Between("hour", hourFrom, hourTo).
		Group("hour").
		FindAll()
	if err != nil {
		return nil, err
	}
	var m = map[string]*NSRecordHourlyStat{} // hour => *NSRecordHourlyStat
	for _, one := range ones {
		m[one.(*NSRecordHourlyStat).Hour] = one.(*NSRecordHourlyStat)
	}
	hours, err := utils.RangeHours(hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, hour := range hours {
		stat, ok := m[hour]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &NSRecordHourlyStat{
				Hour: hour,
			})
		}
	}
	return
}

// FindDailyStats 按天统计
func (this *NSRecordHourlyStatDAO) FindDailyStats(tx *dbs.Tx, dayFrom string, dayTo string) (result []*NSRecordHourlyStat, err error) {
	ones, err := this.Query(tx).
		Result("day", "SUM(countRequests) AS countRequests", "SUM(bytes) AS bytes").
		Between("day", dayFrom, dayTo).
		Group("day").
		FindAll()
	if err != nil {
		return nil, err
	}
	var m = map[string]*NSRecordHourlyStat{} // day => *NSRecordHourlyStat
	for _, one := range ones {
		m[one.(*NSRecordHourlyStat).Day] = one.(*NSRecordHourlyStat)
	}
	days, err := utils.RangeDays(dayFrom, dayTo)
	if err != nil {
		return nil, err
	}
	for _, day := range days {
		stat, ok := m[day]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &NSRecordHourlyStat{
				Day: day,
			})
		}
	}
	return
}

// ListTopNodes 节点排行
func (this *NSRecordHourlyStatDAO) ListTopNodes(tx *dbs.Tx, hourFrom string, hourTo string, size int64) (result []*NSRecordHourlyStat, err error) {
	_, err = this.Query(tx).
		Result("MIN(clusterId) AS clusterId", "nodeId", "SUM(countRequests) AS countRequests", "SUM(bytes) AS bytes").
		Between("hour", hourFrom, hourTo).
		Group("nodeId").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// ListTopDomains 域名排行
func (this *NSRecordHourlyStatDAO) ListTopDomains(tx *dbs.Tx, hourFrom string, hourTo string, size int64) (result []*NSRecordHourlyStat, err error) {
	_, err = this.Query(tx).
		Result("domainId", "SUM(countRequests) AS countRequests", "SUM(bytes) AS bytes").
		Between("hour", hourFrom, hourTo).
		Group("domainId").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// Clean 清理历史数据
func (this *NSRecordHourlyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var hour = timeutil.Format("Ymd00", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("hour", hour).
		Delete()
	return err
}
