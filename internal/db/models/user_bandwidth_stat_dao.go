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
	"strings"
	"sync"
	"time"
)

type UserBandwidthStatDAO dbs.DAO

const (
	UserBandwidthStatTablePartials = 20
)

var fullDataMap = map[string]bool{} // month => bool
var fullDataLocker = &sync.Mutex{}

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedUserBandwidthStatDAO.CleanDefaultDays(nil, 100)
				if err != nil {
					remotelogs.Error("SharedUserBandwidthStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

func NewUserBandwidthStatDAO() *UserBandwidthStatDAO {
	return dbs.NewDAO(&UserBandwidthStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserBandwidthStats",
			Model:  new(UserBandwidthStat),
			PkName: "id",
		},
	}).(*UserBandwidthStatDAO)
}

var SharedUserBandwidthStatDAO *UserBandwidthStatDAO

func init() {
	dbs.OnReady(func() {
		SharedUserBandwidthStatDAO = NewUserBandwidthStatDAO()
	})
}

// UpdateUserBandwidth 写入数据
func (this *UserBandwidthStatDAO) UpdateUserBandwidth(tx *dbs.Tx, userId int64, regionId int64, day string, timeAt string, bytes int64, totalBytes int64, cachedBytes int64, attackBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64) error {
	if userId <= 0 {
		// 如果用户ID不大于0，则说明服务不属于任何用户，此时不需要处理
		return nil
	}

	return this.Query(tx).
		Table(this.partialTable(userId)).
		Param("bytes", bytes).
		Param("totalBytes", totalBytes).
		Param("cachedBytes", cachedBytes).
		Param("attackBytes", attackBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		Param("countAttackRequests", countAttackRequests).
		InsertOrUpdateQuickly(maps.Map{
			"userId":              userId,
			"regionId":            regionId,
			"day":                 day,
			"timeAt":              timeAt,
			"bytes":               bytes,
			"totalBytes":          totalBytes,
			"avgBytes":            totalBytes / 300,
			"cachedBytes":         cachedBytes,
			"attackBytes":         attackBytes,
			"countRequests":       countRequests,
			"countCachedRequests": countCachedRequests,
			"countAttackRequests": countAttackRequests,
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

// FindUserPeekBandwidthInMonth 读取某月带宽峰值
// month YYYYMM
func (this *UserBandwidthStatDAO) FindUserPeekBandwidthInMonth(tx *dbs.Tx, userId int64, month string, useAvg bool) (*UserBandwidthStat, error) {
	one, err := this.Query(tx).
		Table(this.partialTable(userId)).
		Result("MIN(id) AS id", "MIN(userId) AS userId", "day", "timeAt", this.sumBytesField(useAvg)).
		Attr("userId", userId).
		Between("day", month+"01", month+"31").
		Desc("bytes").
		Group("day").
		Group("timeAt").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*UserBandwidthStat), nil
}

// FindPercentileBetweenDays 获取日期段内内百分位
// regionId 如果为 -1 表示没有区域的带宽；如果为 0 表示所有区域的带宽
func (this *UserBandwidthStatDAO) FindPercentileBetweenDays(tx *dbs.Tx, userId int64, regionId int64, dayFrom string, dayTo string, percentile int32, useAvg bool) (result *UserBandwidthStat, err error) {
	if dayFrom > dayTo {
		dayFrom, dayTo = dayTo, dayFrom
	}

	if percentile <= 0 {
		percentile = 95
	}

	// 如果是100%以上，则快速返回
	if percentile >= 100 {
		var query = this.Query(tx).
			Table(this.partialTable(userId))
		if regionId > 0 {
			query.Attr("regionId", regionId)
		} else if regionId < 0 {
			query.Attr("regionId", 0)
		}
		one, err := query.
			Result("MIN(id) AS id", "MIN(userId) AS userId", "day", "timeAt", this.sumBytesField(useAvg)).
			Attr("userId", userId).
			Between("day", dayFrom, dayTo).
			Desc("bytes").
			Group("day").
			Group("timeAt").
			Find()
		if err != nil || one == nil {
			return nil, err
		}

		return one.(*UserBandwidthStat), nil
	}

	// 总数量
	var totalQuery = this.Query(tx).
		Table(this.partialTable(userId))
	if regionId > 0 {
		totalQuery.Attr("regionId", regionId)
	} else if regionId < 0 {
		totalQuery.Attr("regionId", 0)
	}
	total, err := totalQuery.
		Attr("userId", userId).
		Between("day", dayFrom, dayTo).
		CountAttr("DISTINCT day, timeAt")
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
	var query = this.Query(tx).
		Table(this.partialTable(userId))
	if regionId > 0 {
		query.Attr("regionId", regionId)
	} else if regionId < 0 {
		query.Attr("regionId", 0)
	}
	one, err := query.
		Result("MIN(id) AS id", "MIN(userId) AS userId", "day", "timeAt", this.sumBytesField(useAvg)).
		Attr("userId", userId).
		Between("day", dayFrom, dayTo).
		Desc("bytes").
		Group("day").
		Group("timeAt").
		Offset(offset).
		Find()
	if err != nil || one == nil {
		return nil, err
	}

	return one.(*UserBandwidthStat), nil
}

// FindUserPeekBandwidthInDay 读取某日带宽峰值
// day YYYYMMDD
func (this *UserBandwidthStatDAO) FindUserPeekBandwidthInDay(tx *dbs.Tx, userId int64, day string, useAvg bool) (*UserBandwidthStat, error) {
	one, err := this.Query(tx).
		Table(this.partialTable(userId)).
		Result("MIN(id) AS id", "MIN(userId) AS userId", "MIN(day) AS day", "timeAt", this.sumBytesField(useAvg)).
		Attr("userId", userId).
		Attr("day", day).
		Desc("bytes").
		Group("timeAt").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*UserBandwidthStat), nil
}

// FindUserBandwidthStatsBetweenDays 查找日期段内的带宽峰值
// dayFrom YYYYMMDD
// dayTo YYYYMMDD
func (this *UserBandwidthStatDAO) FindUserBandwidthStatsBetweenDays(tx *dbs.Tx, userId int64, regionId int64, dayFrom string, dayTo string, useAvg bool) (result []*pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat, err error) {
	if userId <= 0 {
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

	var query = this.Query(tx).
		Table(this.partialTable(userId))
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	ones, _, err := query.
		Result(this.sumBytesField(useAvg), "day", "timeAt").
		Attr("userId", userId).
		Between("day", dayFrom, dayTo).
		Group("day").
		Group("timeAt").
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

// FindDistinctUserIds 获取所有有带宽的用户ID
// dayFrom YYYYMMDD
// dayTo YYYYMMDD
func (this *UserBandwidthStatDAO) FindDistinctUserIds(tx *dbs.Tx, dayFrom string, dayTo string) (userIds []int64, err error) {
	dayFrom = strings.ReplaceAll(dayFrom, "-", "")
	dayTo = strings.ReplaceAll(dayTo, "-", "")

	err = this.runBatch(func(table string, locker *sync.Mutex) error {
		ones, err := this.Query(tx).
			Table(table).
			Between("day", dayFrom, dayTo).
			Result("DISTINCT userId").
			FindAll()
		if err != nil {
			return err
		}

		for _, one := range ones {
			locker.Lock()
			var userId = int64(one.(*UserBandwidthStat).UserId)
			if userId > 0 {
				userIds = append(userIds, userId)
			}
			locker.Unlock()
		}
		return nil
	})
	return
}

// SumUserMonthly 获取某月流量总和
// month 格式为YYYYMM
func (this *UserBandwidthStatDAO) SumUserMonthly(tx *dbs.Tx, userId int64, month string) (int64, error) {
	// 兼容以往版本
	hasFullData, err := this.HasFullData(tx, userId, month)
	if err != nil {
		return 0, err
	}
	if !hasFullData {
		return SharedServerDailyStatDAO.compatSumUserMonthly(tx, userId, month)
	}

	return this.Query(tx).
		Table(this.partialTable(userId)).
		Between("day", month+"01", month+"31").
		Attr("userId", userId).
		SumInt64("totalBytes", 0)
}

// SumUserDaily 获取某天流量总和
// day 格式为YYYYMMDD
func (this *UserBandwidthStatDAO) SumUserDaily(tx *dbs.Tx, userId int64, regionId int64, day string) (stat *UserBandwidthStat, err error) {
	if !regexputils.YYYYMMDD.MatchString(day) {
		return nil, nil
	}

	// 兼容以往版本
	hasFullData, err := this.HasFullData(tx, userId, day[:6])
	if err != nil {
		return nil, err
	}
	if !hasFullData {
		serverStat, err := SharedServerDailyStatDAO.compatSumUserDaily(tx, userId, regionId, day)
		if err != nil || serverStat == nil {
			return nil, err
		}

		return serverStat.AsUserBandwidthStat(), nil
	}

	var query = this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}

	one, err := query.
		Table(this.partialTable(userId)).
		Attr("day", day).
		Attr("userId", userId).
		Result("SUM(totalBytes) AS totalBytes", "SUM(cachedBytes) AS cachedBytes", "SUM(attackBytes) AS attackBytes", "SUM(countRequests) AS countRequests", "SUM(countCachedRequests) AS countCachedRequests", "SUM(countAttackRequests) AS countAttackRequests").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*UserBandwidthStat), nil
}

// SumDailyStat 获取某天内的流量
// dayFrom 格式为YYYYMMDD
// dayTo 格式为YYYYMMDD
func (this *UserBandwidthStatDAO) SumDailyStat(tx *dbs.Tx, userId int64, regionId int64, dayFrom string, dayTo string) (stat *pb.ServerDailyStat, err error) {
	if !regexputils.YYYYMMDD.MatchString(dayFrom) {
		return nil, errors.New("invalid dayFrom '" + dayFrom + "'")
	}
	if !regexputils.YYYYMMDD.MatchString(dayTo) {
		return nil, errors.New("invalid dayTo '" + dayTo + "'")
	}

	// 兼容以往版本
	hasFullData, err := this.HasFullData(tx, userId, dayFrom[:6])
	if err != nil {
		return nil, err
	}
	if !hasFullData {
		return SharedServerDailyStatDAO.compatSumDailyStat(tx, userId, 0, regionId, dayFrom, dayTo)
	}

	stat = &pb.ServerDailyStat{}

	if userId <= 0 {
		return
	}

	if dayFrom > dayTo {
		dayFrom, dayTo = dayTo, dayFrom
	}

	var query = this.Query(tx).
		Table(this.partialTable(userId)).
		Result("SUM(totalBytes) AS totalBytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes")

	query.Attr("userId", userId)

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

// CleanDays 清理过期数据
func (this *UserBandwidthStatDAO) CleanDays(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days)) // 保留大约3个月的数据
	return this.runBatch(func(table string, locker *sync.Mutex) error {
		_, err := this.Query(tx).
			Table(table).
			Lt("day", day).
			Delete()
		return err
	})
}

func (this *UserBandwidthStatDAO) CleanDefaultDays(tx *dbs.Tx, defaultDays int) error {
	databaseConfig, err := SharedSysSettingDAO.ReadDatabaseConfig(tx)
	if err != nil {
		return err
	}

	if databaseConfig != nil && databaseConfig.UserBandwidthStat.Clean.Days > 0 {
		defaultDays = databaseConfig.UserBandwidthStat.Clean.Days
	}
	if defaultDays <= 0 {
		defaultDays = 100
	}

	return this.CleanDays(tx, defaultDays)
}

// 批量执行
func (this *UserBandwidthStatDAO) runBatch(f func(table string, locker *sync.Mutex) error) error {
	var locker = &sync.Mutex{}
	var wg = sync.WaitGroup{}
	wg.Add(UserBandwidthStatTablePartials)
	var resultErr error
	for i := 0; i < UserBandwidthStatTablePartials; i++ {
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
func (this *UserBandwidthStatDAO) partialTable(userId int64) string {
	return this.Table + "_" + types.String(userId%int64(UserBandwidthStatTablePartials))
}

// 获取总数字段
func (this *UserBandwidthStatDAO) sumBytesField(useAvg bool) string {
	if useAvg {
		return "SUM(avgBytes) AS bytes"
	}
	return "SUM(bytes) AS bytes"
}

func (this *UserBandwidthStatDAO) fixUserStat(stat *UserBandwidthStat, useAvg bool) *UserBandwidthStat {
	if stat == nil {
		return nil
	}
	if useAvg {
		stat.Bytes = stat.AvgBytes
	}
	return stat
}

// HasFullData 检查一个月是否完整数据
// 是为了兼容以前数据，以前的表中没有缓存流量、请求数等字段
func (this *UserBandwidthStatDAO) HasFullData(tx *dbs.Tx, userId int64, month string) (bool, error) {
	if !regexputils.YYYYMM.MatchString(month) {
		return false, errors.New("invalid month '" + month + "'")
	}

	fullDataLocker.Lock()
	hasData, ok := fullDataMap[month]
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
		Table(this.partialTable(userId)).
		Between("day", lastMonthString+"01", lastMonthString+"31").
		DescPk().
		Find()
	if err != nil {
		return false, err
	}

	var b = one != nil && one.(*UserBandwidthStat).CountRequests > 0
	fullDataLocker.Lock()
	fullDataMap[month] = b
	fullDataLocker.Unlock()

	return b, nil
}
