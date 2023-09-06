package models

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/regexputils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ServerBandwidthStatDAO dbs.DAO

const (
	ServerBandwidthStatTablePartitions = 20 // 分表数量
)

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedServerBandwidthStatDAO.CleanDefaultDays(nil, 100)
				if err != nil {
					remotelogs.Error("SharedServerBandwidthStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

func NewServerBandwidthStatDAO() *ServerBandwidthStatDAO {
	return dbs.NewDAO(&ServerBandwidthStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerBandwidthStats",
			Model:  new(ServerBandwidthStat),
			PkName: "id",
		},
	}).(*ServerBandwidthStatDAO)
}

var SharedServerBandwidthStatDAO *ServerBandwidthStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerBandwidthStatDAO = NewServerBandwidthStatDAO()
	})
}

// UpdateServerBandwidth 写入数据
// 现在不需要把 userPlanId 加入到数据表unique key中，因为只会影响5分钟统计，影响非常有限
func (this *ServerBandwidthStatDAO) UpdateServerBandwidth(tx *dbs.Tx, userId int64, serverId int64, regionId int64, userPlanId int64, day string, timeAt string, bandwidthBytes int64, totalBytes int64, cachedBytes int64, attackBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64) error {
	if serverId <= 0 {
		return errors.New("invalid server id '" + types.String(serverId) + "'")
	}

	return this.Query(tx).
		Table(this.partialTable(serverId)).
		Param("bytes", bandwidthBytes).
		Param("totalBytes", totalBytes).
		Param("cachedBytes", cachedBytes).
		Param("attackBytes", attackBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		Param("countAttackRequests", countAttackRequests).
		InsertOrUpdateQuickly(maps.Map{
			"userId":              userId,
			"serverId":            serverId,
			"regionId":            regionId,
			"day":                 day,
			"timeAt":              timeAt,
			"bytes":               bandwidthBytes,
			"totalBytes":          totalBytes,
			"avgBytes":            totalBytes / 300,
			"cachedBytes":         cachedBytes,
			"attackBytes":         attackBytes,
			"countRequests":       countRequests,
			"countCachedRequests": countCachedRequests,
			"countAttackRequests": countAttackRequests,
			"userPlanId":          userPlanId,
		}, maps.Map{
			"bytes":               dbs.SQL("bytes+:bytes"),
			"avgBytes":            dbs.SQL("(totalBytes+:totalBytes)/300"), // 因为生成SQL语句时会自动将avgBytes排在totalBytes之前，所以这里不用担心先后顺序的问题
			"totalBytes":          dbs.SQL("totalBytes+:totalBytes"),
			"cachedBytes":         dbs.SQL("cachedBytes+:cachedBytes"),
			"attackBytes":         dbs.SQL("attackBytes+:attackBytes"),
			"countRequests":       dbs.SQL("countRequests+:countRequests"),
			"countCachedRequests": dbs.SQL("countCachedRequests+:countCachedRequests"),
			"countAttackRequests": dbs.SQL("countAttackRequests+:countAttackRequests"),
		})
}

// FindMinutelyPeekBandwidthBytes 获取某分钟的带宽峰值
// day YYYYMMDD
// minute HHII
func (this *ServerBandwidthStatDAO) FindMinutelyPeekBandwidthBytes(tx *dbs.Tx, serverId int64, day string, minute string, useAvg bool) (int64, error) {
	return this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Result(this.bytesField(useAvg)).
		Attr("day", day).
		Attr("timeAt", minute).
		FindInt64Col(0)
}

// FindHourlyBandwidthStats 按小时获取带宽峰值
func (this *ServerBandwidthStatDAO) FindHourlyBandwidthStats(tx *dbs.Tx, serverId int64, hours int32, useAvg bool) (result []*pb.FindHourlyServerBandwidthStatsResponse_Stat, err error) {
	if hours <= 0 {
		hours = 24
	}

	var timestamp = time.Now().Unix() - int64(hours)*3600

	ones, _, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Between("day", timeutil.FormatTime("Ymd", timestamp), timeutil.Format("Ymd")).
		Attr("serverId", serverId).
		Result(this.maxBytesField(useAvg), "CONCAT(day, '.', SUBSTRING(timeAt, 1, 2)) AS fullTime").
		Gte("CONCAT(day, '.', SUBSTRING(timeAt, 1, 2))", timeutil.FormatTime("Ymd.H", timestamp)).
		Group("fullTime").
		FindOnes()
	if err != nil {
		return nil, err
	}

	var m = map[string]*pb.FindHourlyServerBandwidthStatsResponse_Stat{}
	for _, one := range ones {
		var fullTime = one.GetString("fullTime")
		var timePieces = strings.Split(fullTime, ".")
		var day = timePieces[0]
		var hour = timePieces[1]
		var bytes = one.GetInt64("bytes")

		m[day+hour] = &pb.FindHourlyServerBandwidthStatsResponse_Stat{
			Bytes: bytes,
			Bits:  bytes * 8,
			Day:   day,
			Hour:  types.Int32(hour),
		}
	}

	fullHours, err := utils.RangeHours(timeutil.FormatTime("YmdH", timestamp), timeutil.Format("YmdH"))
	if err != nil {
		return nil, err
	}
	for _, fullHour := range fullHours {
		stat, ok := m[fullHour]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &pb.FindHourlyServerBandwidthStatsResponse_Stat{
				Bytes: 0,
				Bits:  0,
				Day:   fullHour[:8],
				Hour:  types.Int32(fullHour[8:]),
			})
		}
	}

	return result, nil
}

// FindDailyPeekBandwidthBytes 获取某天的带宽峰值
// day YYYYMMDD
func (this *ServerBandwidthStatDAO) FindDailyPeekBandwidthBytes(tx *dbs.Tx, serverId int64, day string, useAvg bool) (int64, error) {
	return this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Attr("day", day).
		Result(this.maxBytesField(useAvg)).
		FindInt64Col(0)
}

// FindDailyBandwidthStats 按天获取带宽峰值
func (this *ServerBandwidthStatDAO) FindDailyBandwidthStats(tx *dbs.Tx, serverId int64, days int32, useAvg bool) (result []*pb.FindDailyServerBandwidthStatsResponse_Stat, err error) {
	if days <= 0 {
		days = 14
	}

	var timestamp = time.Now().Unix() - int64(days)*86400
	ones, _, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Result(this.maxBytesField(useAvg), "day").
		Attr("serverId", serverId).
		Gte("day", timeutil.FormatTime("Ymd", timestamp)).
		Group("day").
		FindOnes()
	if err != nil {
		return nil, err
	}

	var m = map[string]*pb.FindDailyServerBandwidthStatsResponse_Stat{}
	for _, one := range ones {
		var day = one.GetString("day")
		var bytes = one.GetInt64("bytes")

		m[day] = &pb.FindDailyServerBandwidthStatsResponse_Stat{
			Bytes: bytes,
			Bits:  bytes * 8,
			Day:   day,
		}
	}

	allDays, err := utils.RangeDays(timeutil.FormatTime("Ymd", timestamp), timeutil.Format("Ymd"))
	if err != nil {
		return nil, err
	}
	for _, day := range allDays {
		stat, ok := m[day]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &pb.FindDailyServerBandwidthStatsResponse_Stat{
				Bytes: 0,
				Bits:  0,
				Day:   day,
			})
		}
	}

	return result, nil
}

// FindBandwidthStatsBetweenDays 查找日期段内的带宽峰值
// dayFrom YYYYMMDD
// dayTo YYYYMMDD
func (this *ServerBandwidthStatDAO) FindBandwidthStatsBetweenDays(tx *dbs.Tx, serverId int64, dayFrom string, dayTo string, useAvg bool) (result []*pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat, err error) {
	if serverId <= 0 {
		return nil, nil
	}

	if !regexputils.YYYYMMDD.MatchString(dayFrom) {
		return nil, errors.New("invalid dayFrom '" + dayFrom + "'")
	}
	if !regexputils.YYYYMMDD.MatchString(dayTo) {
		return nil, errors.New("invalid dayTo '" + dayTo + "'")
	}

	if dayFrom > dayTo {
		dayFrom, dayTo = dayTo, dayFrom
	}

	ones, _, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Result(this.bytesField(useAvg), "day", "timeAt").
		Attr("serverId", serverId).
		Between("day", dayFrom, dayTo).
		FindOnes()
	if err != nil {
		return nil, err
	}

	var m = map[string]*pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat{}
	for _, one := range ones {
		var day = one.GetString("day")
		var bytes = one.GetInt64("bytes")
		var timeAt = one.GetString("timeAt")
		var key = day + "@" + timeAt

		m[key] = &pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat{
			Bytes:  bytes,
			Bits:   bytes * 8,
			Day:    day,
			TimeAt: timeAt,
		}
	}

	allDays, err := utils.RangeDays(dayFrom, dayTo)
	if err != nil {
		return nil, err
	}

	dayTimes, err := utils.Range24HourTimes(5)
	if err != nil {
		return nil, err
	}

	// 截止到当前时间
	var currentTime = timeutil.Format("Ymd@Hi")

	for _, day := range allDays {
		for _, timeAt := range dayTimes {
			var key = day + "@" + timeAt
			if key >= currentTime {
				break
			}

			stat, ok := m[key]
			if ok {
				result = append(result, stat)
			} else {
				result = append(result, &pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat{
					Day:    day,
					TimeAt: timeAt,
				})
			}
		}
	}

	return result, nil
}

// FindMonthlyPeekBandwidthBytes 获取某月的带宽峰值
// month YYYYMM
func (this *ServerBandwidthStatDAO) FindMonthlyPeekBandwidthBytes(tx *dbs.Tx, serverId int64, month string, useAvg bool) (int64, error) {
	return this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Between("day", month+"01", month+"31").
		Result(this.maxBytesField(useAvg)).
		FindInt64Col(0)
}

// FindServerStats 查找某个时间段的带宽统计
// 参数：
//   - day YYYYMMDD
//   - timeAt HHII
func (this *ServerBandwidthStatDAO) FindServerStats(tx *dbs.Tx, serverId int64, day string, timeFrom string, timeTo string, useAvg bool) (result []*ServerBandwidthStat, err error) {
	_, err = this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Attr("day", day).
		Between("timeAt", timeFrom, timeTo).
		Slice(&result).
		FindAll()

	// 使用平均带宽
	this.fixServerStats(result, useAvg)

	return
}

// FindAllServerStatsWithDay 查找某个服务的当天的所有带宽峰值
// day YYYYMMDD
func (this *ServerBandwidthStatDAO) FindAllServerStatsWithDay(tx *dbs.Tx, serverId int64, day string, useAvg bool) (result []*ServerBandwidthStat, err error) {
	_, err = this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Attr("day", day).
		AscPk().
		Slice(&result).
		FindAll()

	// 使用平均带宽
	this.fixServerStats(result, useAvg)

	return
}

// FindAllServerStatsWithMonth 查找某个服务的当月的所有带宽峰值
// month YYYYMM
func (this *ServerBandwidthStatDAO) FindAllServerStatsWithMonth(tx *dbs.Tx, serverId int64, month string, useAvg bool) (result []*ServerBandwidthStat, err error) {
	_, err = this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Between("day", month+"01", month+"31").
		AscPk().
		Slice(&result).
		FindAll()

	// 使用平均带宽
	this.fixServerStats(result, useAvg)

	return
}

// FindMonthlyPercentile 获取某月内百分位
func (this *ServerBandwidthStatDAO) FindMonthlyPercentile(tx *dbs.Tx, serverId int64, month string, percentile int, useAvg bool, noPlan bool) (result int64, err error) {
	if percentile <= 0 {
		percentile = 95
	}

	// 如果是100%以上，则快速返回
	if percentile >= 100 {
		var query = this.Query(tx)
		if noPlan {
			query.Attr("userPlanId", 0)
		}
		result, err = query.
			Table(this.partialTable(serverId)).
			Attr("serverId", serverId).
			Result(this.bytesField(useAvg)).
			Between("day", month+"01", month+"31").
			Desc("bytes").
			Limit(1).
			FindInt64Col(0)
		return
	}

	// 总数量
	var totalQuery = this.Query(tx)
	if noPlan {
		totalQuery.Attr("userPlanId", 0)
	}
	total, err := totalQuery.
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Between("day", month+"01", month+"31").
		Count()
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}

	var offset int64

	if total > 1 {
		offset = int64(math.Ceil(float64(total) * float64(100-percentile) / 100))
	}

	// 查询 nth 位置
	var query = this.Query(tx)
	if noPlan {
		query.Attr("userPlanId", 0)
	}
	result, err = query.
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Result(this.bytesField(useAvg)).
		Between("day", month+"01", month+"31").
		Desc("bytes").
		Offset(offset).
		Limit(1).
		FindInt64Col(0)

	return
}

// FindPercentileBetweenDays 获取日期段内内百分位
func (this *ServerBandwidthStatDAO) FindPercentileBetweenDays(tx *dbs.Tx, serverId int64, dayFrom string, dayTo string, percentile int32, useAvg bool) (result *ServerBandwidthStat, err error) {
	if dayFrom > dayTo {
		dayFrom, dayTo = dayTo, dayFrom
	}

	if percentile <= 0 {
		percentile = 95
	}

	// 如果是100%以上，则快速返回
	if percentile >= 100 {
		one, err := this.Query(tx).
			Table(this.partialTable(serverId)).
			Attr("serverId", serverId).
			Between("day", dayFrom, dayTo).
			Desc(this.bytesOrderField(useAvg)).
			Find()
		if err != nil || one == nil {
			return nil, err
		}

		return this.fixServerStat(one.(*ServerBandwidthStat), useAvg), nil
	}

	// 总数量
	total, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Between("day", dayFrom, dayTo).
		Count()
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}

	var offset int64

	if total > 1 {
		offset = int64(math.Ceil(float64(total) * float64(100-percentile) / 100))
	}

	// 查询 nth 位置
	one, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Between("day", dayFrom, dayTo).
		Desc(this.bytesOrderField(useAvg)).
		Offset(offset).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return this.fixServerStat(one.(*ServerBandwidthStat), useAvg), nil
}

// FindPercentileBetweenTimes 获取时间段内内百分位
// timeFrom 开始时间，格式 YYYYMMDDHHII
// timeTo 结束时间，格式 YYYYMMDDHHII
func (this *ServerBandwidthStatDAO) FindPercentileBetweenTimes(tx *dbs.Tx, serverId int64, timeFrom string, timeTo string, percentile int32, useAvg bool) (result *ServerBandwidthStat, err error) {
	var reg = regexp.MustCompile(`^\d{12}$`)
	if !reg.MatchString(timeFrom) {
		return nil, errors.New("invalid timeFrom '" + timeFrom + "'")
	}
	if !reg.MatchString(timeTo) {
		return nil, errors.New("invalid timeTo '" + timeTo + "'")
	}

	if timeFrom > timeTo {
		timeFrom, timeTo = timeTo, timeFrom
	}

	if percentile <= 0 {
		percentile = 95
	}

	// 如果是100%以上，则快速返回
	if percentile >= 100 {
		one, err := this.Query(tx).
			Table(this.partialTable(serverId)).
			Attr("serverId", serverId).
			Between("CONCAT(day, timeAt)", timeFrom, timeTo).
			Desc(this.bytesOrderField(useAvg)).
			Find()
		if err != nil || one == nil {
			return nil, err
		}

		return this.fixServerStat(one.(*ServerBandwidthStat), useAvg), nil
	}

	// 总数量
	total, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Between("day", timeFrom[:8], timeTo[:8]).
		Attr("serverId", serverId).
		Between("CONCAT(day, timeAt)", timeFrom, timeTo).
		Count()
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}

	var offset int64

	if total > 1 {
		offset = int64(math.Ceil(float64(total) * float64(100-percentile) / 100))
	}

	// 查询 nth 位置
	one, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Between("day", timeFrom[:8], timeTo[:8]).
		Attr("serverId", serverId).
		Between("CONCAT(day, timeAt)", timeFrom, timeTo).
		Desc(this.bytesOrderField(useAvg)).
		Offset(offset).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return this.fixServerStat(one.(*ServerBandwidthStat), useAvg), nil
}

// FindDailyStats 按天统计
func (this *ServerBandwidthStatDAO) FindDailyStats(tx *dbs.Tx, serverId int64, dayFrom string, dayTo string) (result []*ServerBandwidthStat, err error) {
	// 兼容以往版本
	if !regexputils.YYYYMMDD.MatchString(dayFrom) || !regexputils.YYYYMMDD.MatchString(dayTo) {
		return nil, nil
	}
	hasFullData, err := this.HasFullData(tx, serverId, dayFrom[:6])
	if err != nil {
		return nil, err
	}
	if !hasFullData {
		ones, err := SharedServerDailyStatDAO.compatFindDailyStats(tx, serverId, dayFrom, dayTo)
		if err != nil {
			return nil, err
		}
		for _, one := range ones {
			result = append(result, one.AsServerBandwidthStat())
		}

		return result, nil
	}

	ones, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Result("SUM(totalBytes) AS totalBytes", "SUM(cachedBytes) AS cachedBytes", "SUM(countRequests) AS countRequests", "SUM(countCachedRequests) AS countCachedRequests", "SUM(countAttackRequests) AS countAttackRequests", "SUM(attackBytes) AS attackBytes", "day").
		Attr("serverId", serverId).
		Between("day", dayFrom, dayTo).
		Group("day").
		FindAll()
	if err != nil {
		return nil, err
	}

	var dayMap = map[string]*ServerBandwidthStat{} // day => Stat
	for _, one := range ones {
		var stat = one.(*ServerBandwidthStat)
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
			result = append(result, &ServerBandwidthStat{Day: day})
		}
	}

	return
}

// FindHourlyStats 按小时统计
func (this *ServerBandwidthStatDAO) FindHourlyStats(tx *dbs.Tx, serverId int64, hourFrom string, hourTo string) (result []*ServerBandwidthStat, err error) {
	// 兼容以往版本
	if !regexputils.YYYYMMDDHH.MatchString(hourFrom) || !regexputils.YYYYMMDDHH.MatchString(hourTo) {
		return nil, nil
	}
	hasFullData, err := this.HasFullData(tx, serverId, hourFrom[:6])
	if err != nil {
		return nil, err
	}
	if !hasFullData {
		ones, err := SharedServerDailyStatDAO.compatFindHourlyStats(tx, serverId, hourFrom, hourTo)
		if err != nil {
			return nil, err
		}
		for _, one := range ones {
			result = append(result, one.AsServerBandwidthStat())
		}

		return result, nil
	}

	var query = this.Query(tx).
		Table(this.partialTable(serverId)).
		Between("day", hourFrom[:8], hourTo[:8]).
		Result("MIN(day) AS day", "MIN(timeAt) AS timeAt", "SUM(totalBytes) AS totalBytes", "SUM(cachedBytes) AS cachedBytes", "SUM(countRequests) AS countRequests", "SUM(countCachedRequests) AS countCachedRequests", "SUM(countAttackRequests) AS countAttackRequests", "SUM(attackBytes) AS attackBytes", "CONCAT(day, SUBSTR(timeAt, 1, 2)) AS hour").
		Attr("serverId", serverId)

	if hourFrom[:8] == hourTo[:8] { // 同一天
		query.Attr("day", hourFrom[:8])
		query.Between("timeAt", hourFrom[8:]+"00", hourTo[8:]+"59")
	} else {
		query.Between("CONCAT(day, SUBSTR(timeAt, 1, 2))", hourFrom, hourTo)
	}

	ones, err := query.
		Group("hour").
		FindAll()
	if err != nil {
		return nil, err
	}

	var hourMap = map[string]*ServerBandwidthStat{} // hour => Stat
	for _, one := range ones {
		var stat = one.(*ServerBandwidthStat)
		var hour = stat.Day + stat.TimeAt[:2]
		hourMap[hour] = stat
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
			result = append(result, &ServerBandwidthStat{
				Day:    hour[:8],
				TimeAt: hour[8:] + "00",
			})
		}
	}

	return
}

// SumDailyStat 获取某天内的流量
// dayFrom 格式为YYYYMMDD
// dayTo 格式为YYYYMMDD
func (this *ServerBandwidthStatDAO) SumDailyStat(tx *dbs.Tx, serverId int64, regionId int64, dayFrom string, dayTo string) (stat *pb.ServerDailyStat, err error) {
	if !regexputils.YYYYMMDD.MatchString(dayFrom) {
		return nil, errors.New("invalid dayFrom '" + dayFrom + "'")
	}
	if !regexputils.YYYYMMDD.MatchString(dayTo) {
		return nil, errors.New("invalid dayTo '" + dayTo + "'")
	}

	// 兼容以往版本
	hasFullData, err := this.HasFullData(tx, serverId, dayFrom[:6])
	if err != nil {
		return nil, err
	}
	if !hasFullData {
		return SharedServerDailyStatDAO.compatSumDailyStat(tx, 0, serverId, regionId, dayFrom, dayTo)
	}

	stat = &pb.ServerDailyStat{}

	if serverId <= 0 {
		return
	}

	if dayFrom > dayTo {
		dayFrom, dayTo = dayTo, dayFrom
	}

	var query = this.Query(tx).
		Table(this.partialTable(serverId)).
		Result("SUM(totalBytes) AS totalBytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes")

	query.Attr("serverId", serverId)

	if regionId > 0 {
		query.Attr("regionId", regionId)
	}

	if dayFrom == dayTo {
		query.Attr("day", dayFrom)
	} else {
		query.Between("day", dayFrom, dayTo)
	}

	one, _, err := query.FindOne()
	if err != nil {
		return nil, err
	}

	if one == nil {
		return
	}

	stat.Bytes = one.GetInt64("totalBytes")
	stat.CachedBytes = one.GetInt64("cachedBytes")
	stat.CountRequests = one.GetInt64("countRequests")
	stat.CountCachedRequests = one.GetInt64("countCachedRequests")
	stat.CountAttackRequests = one.GetInt64("countAttackRequests")
	stat.AttackBytes = one.GetInt64("attackBytes")
	return
}

// SumMonthlyBytes 统计某个网站单月总流量
func (this *ServerBandwidthStatDAO) SumMonthlyBytes(tx *dbs.Tx, serverId int64, month string, noPlan bool) (int64, error) {
	if !regexputils.YYYYMM.MatchString(month) {
		return 0, errors.New("invalid month '" + month + "'")
	}

	// 兼容以往版本
	hasFullData, err := this.HasFullData(tx, serverId, month)
	if err != nil {
		return 0, err
	}
	if !hasFullData {
		return SharedServerDailyStatDAO.SumMonthlyBytes(tx, serverId, month)
	}

	var query = this.Query(tx)
	if noPlan {
		query.Attr("userPlanId", 0)
	}
	return query.
		Table(this.partialTable(serverId)).
		Between("day", month+"01", month+"31").
		Attr("serverId", serverId).
		SumInt64("totalBytes", 0)
}

// SumServerMonthlyWithRegion 根据服务计算某月合计
// month 格式为YYYYMM
func (this *ServerBandwidthStatDAO) SumServerMonthlyWithRegion(tx *dbs.Tx, serverId int64, regionId int64, month string, noPlan bool) (int64, error) {
	var query = this.Query(tx)
	query.Table(this.partialTable(serverId))
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	if noPlan {
		query.Attr("userPlanId", 0)
	}
	return query.Between("day", month+"01", month+"31").
		Attr("serverId", serverId).
		SumInt64("totalBytes", 0)
}

// FindDistinctServerIdsWithoutPlanAtPartition 查找没有绑定套餐的有流量网站
func (this *ServerBandwidthStatDAO) FindDistinctServerIdsWithoutPlanAtPartition(tx *dbs.Tx, partitionIndex int, month string) (serverIds []int64, err error) {
	ones, err := this.Query(tx).
		Table(this.partialTable(int64(partitionIndex))).
		Between("day", month+"01", month+"31").
		Attr("userPlanId", 0). // 没有绑定套餐
		Result("DISTINCT serverId").
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		var serverId = int64(one.(*ServerBandwidthStat).ServerId)
		if serverId <= 0 {
			continue
		}
		serverIds = append(serverIds, serverId)
	}
	return
}

// CountPartitions 查看分区数量
func (this *ServerBandwidthStatDAO) CountPartitions() int {
	return ServerBandwidthStatTablePartitions
}

// CleanDays 清理过期数据
func (this *ServerBandwidthStatDAO) CleanDays(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days)) // 保留大约3个月的数据
	return this.runBatch(func(table string, locker *sync.Mutex) error {
		_, err := this.Query(tx).
			Table(table).
			Lt("day", day).
			Delete()
		return err
	})
}

func (this *ServerBandwidthStatDAO) CleanDefaultDays(tx *dbs.Tx, defaultDays int) error {
	databaseConfig, err := SharedSysSettingDAO.ReadDatabaseConfig(tx)
	if err != nil {
		return err
	}

	if databaseConfig != nil && databaseConfig.ServerBandwidthStat.Clean.Days > 0 {
		defaultDays = databaseConfig.ServerBandwidthStat.Clean.Days
	}
	if defaultDays <= 0 {
		defaultDays = 100
	}

	return this.CleanDays(tx, defaultDays)
}

// 批量执行
func (this *ServerBandwidthStatDAO) runBatch(f func(table string, locker *sync.Mutex) error) error {
	var locker = &sync.Mutex{}
	var wg = sync.WaitGroup{}
	wg.Add(ServerBandwidthStatTablePartitions)
	var resultErr error
	for i := 0; i < ServerBandwidthStatTablePartitions; i++ {
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

// 获取分区表
func (this *ServerBandwidthStatDAO) partialTable(serverId int64) string {
	return this.Table + "_" + types.String(serverId%int64(ServerBandwidthStatTablePartitions))
}

// 获取字节字段
func (this *ServerBandwidthStatDAO) bytesField(useAvg bool) string {
	if useAvg {
		return "avgBytes AS bytes"
	}
	return "bytes"
}

// 获取最大字节字段
func (this *ServerBandwidthStatDAO) maxBytesField(useAvg bool) string {
	if useAvg {
		return "MAX(avgBytes) AS bytes"
	}
	return "MAX(bytes) AS bytes"
}

// 获取排序字段
func (this *ServerBandwidthStatDAO) bytesOrderField(useAvg bool) string {
	if useAvg {
		return "avgBytes"
	}
	return "bytes"
}

func (this *ServerBandwidthStatDAO) fixServerStat(stat *ServerBandwidthStat, useAvg bool) *ServerBandwidthStat {
	if stat == nil {
		return nil
	}
	if useAvg {
		stat.Bytes = stat.AvgBytes
	}
	return stat
}

func (this *ServerBandwidthStatDAO) fixServerStats(stats []*ServerBandwidthStat, useAvg bool) {
	if useAvg {
		for _, stat := range stats {
			stat.Bytes = stat.AvgBytes
		}
	}
}

// HasFullData 检查一个月是否完整数据
// 是为了兼容以前数据，以前的表中没有缓存流量、请求数等字段
func (this *ServerBandwidthStatDAO) HasFullData(tx *dbs.Tx, serverId int64, month string) (bool, error) {
	// 最迟在2024年完成过渡
	if time.Now().Year() >= 2024 {
		return true, nil
	}

	var monthKey = month + "@" + types.String(serverId)

	if !regexputils.YYYYMM.MatchString(month) {
		return false, errors.New("invalid month '" + month + "'")
	}

	fullDataLocker.Lock()
	hasData, ok := fullDataMap[monthKey]
	fullDataLocker.Unlock()
	if ok {
		return hasData, nil
	}

	var year = types.Int(month[:4])
	var monthInt = types.Int(month[4:])

	if year < 2000 || monthInt > 12 || monthInt < 1 {
		return false, nil
	}

	var lastMonth = monthInt - 1
	if lastMonth == 0 {
		lastMonth = 12
		year--
	}

	var lastMonthString = fmt.Sprintf("%d%02d", year, lastMonth)
	one, err := this.Query(tx).
		Table(this.partialTable(serverId)).
		Between("day", lastMonthString+"01", lastMonthString+"31").
		DescPk().
		Find()
	if err != nil {
		return false, err
	}

	var b = one != nil && one.(*ServerBandwidthStat).CountRequests > 0
	fullDataLocker.Lock()
	fullDataMap[monthKey] = b
	fullDataLocker.Unlock()

	return b, nil
}
