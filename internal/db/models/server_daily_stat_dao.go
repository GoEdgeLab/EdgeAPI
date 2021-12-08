package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
	"time"
)

type ServerDailyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedServerDailyStatDAO.Clean(nil, 30) // 只保留N天
				if err != nil {
					logs.Println("ServerDailyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

func NewServerDailyStatDAO() *ServerDailyStatDAO {
	return dbs.NewDAO(&ServerDailyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerDailyStats",
			Model:  new(ServerDailyStat),
			PkName: "id",
		},
	}).(*ServerDailyStatDAO)
}

var SharedServerDailyStatDAO *ServerDailyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerDailyStatDAO = NewServerDailyStatDAO()
	})
}

// SaveStats 提交数据
func (this *ServerDailyStatDAO) SaveStats(tx *dbs.Tx, stats []*pb.ServerDailyStat) error {
	var serverUserMap = map[int64]int64{} // serverId => userId
	var cacheMap = utils.NewCacheMap()
	for _, stat := range stats {
		day := timeutil.FormatTime("Ymd", stat.CreatedAt)
		hour := timeutil.FormatTime("YmdH", stat.CreatedAt)
		timeFrom := timeutil.FormatTime("His", stat.CreatedAt)
		timeTo := timeutil.FormatTime("His", stat.CreatedAt+5*60-1) // 5分钟

		// 所属用户
		serverUserId, ok := serverUserMap[stat.ServerId]
		if !ok {
			userId, err := SharedServerDAO.FindServerUserId(tx, stat.ServerId)
			if err != nil {
				return err
			}
			serverUserId = userId
		}

		_, _, err := this.Query(tx).
			Param("bytes", stat.Bytes).
			Param("cachedBytes", stat.CachedBytes).
			Param("countRequests", stat.CountRequests).
			Param("countCachedRequests", stat.CountCachedRequests).
			Param("countAttackRequests", stat.CountAttackRequests).
			Param("attackBytes", stat.AttackBytes).
			InsertOrUpdate(maps.Map{
				"userId":              serverUserId,
				"serverId":            stat.ServerId,
				"regionId":            stat.RegionId,
				"bytes":               stat.Bytes,
				"cachedBytes":         stat.CachedBytes,
				"countRequests":       stat.CountRequests,
				"countCachedRequests": stat.CountCachedRequests,
				"countAttackRequests": stat.CountAttackRequests,
				"attackBytes":         stat.AttackBytes,
				"planId":              stat.PlanId,
				"day":                 day,
				"hour":                hour,
				"timeFrom":            timeFrom,
				"timeTo":              timeTo,
			}, maps.Map{
				"bytes":               dbs.SQL("bytes+:bytes"),
				"cachedBytes":         dbs.SQL("cachedBytes+:cachedBytes"),
				"countRequests":       dbs.SQL("countRequests+:countRequests"),
				"countCachedRequests": dbs.SQL("countCachedRequests+:countCachedRequests"),
				"countAttackRequests": dbs.SQL("countAttackRequests+:countAttackRequests"),
				"attackBytes":         dbs.SQL("attackBytes+:attackBytes"),
				"planId":              stat.PlanId,
			})
		if err != nil {
			return err
		}

		// 更新流量限制状态
		if stat.CheckTrafficLimiting {
			trafficLimitConfig, err := SharedServerDAO.CalculateServerTrafficLimitConfig(tx, stat.ServerId, cacheMap)
			if err != nil {
				return err
			}
			if trafficLimitConfig != nil && trafficLimitConfig.IsOn && !trafficLimitConfig.IsEmpty() {
				err = SharedServerDAO.IncreaseServerTotalTraffic(tx, stat.ServerId, stat.Bytes)
				if err != nil {
					return err
				}

				err = SharedServerDAO.UpdateServerTrafficLimitStatus(tx, trafficLimitConfig, stat.ServerId, false)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// SumUserMonthly 根据用户计算某月合计
// month 格式为YYYYMM
func (this *ServerDailyStatDAO) SumUserMonthly(tx *dbs.Tx, userId int64, regionId int64, month string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	return query.Between("day", month+"01", month+"32").
		Attr("userId", userId).
		SumInt64("bytes", 0)
}

// SumUserMonthlyWithoutPlan 根据用户计算某月合计并排除套餐
// month 格式为YYYYMM
func (this *ServerDailyStatDAO) SumUserMonthlyWithoutPlan(tx *dbs.Tx, userId int64, regionId int64, month string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	return query.
		Attr("planId", 0).
		Between("day", month+"01", month+"32").
		Attr("userId", userId).
		SumInt64("bytes", 0)
}

// SumUserMonthlyFee 计算用户某个月费用
// month 格式为YYYYMM
func (this *ServerDailyStatDAO) SumUserMonthlyFee(tx *dbs.Tx, userId int64, month string) (float64, error) {
	return this.Query(tx).
		Attr("userId", userId).
		Between("day", month+"01", month+"32").
		Gt("fee", 0).
		Sum("fee", 0)
}

// SumUserMonthlyPeek 获取某月带宽峰值
// month 格式为YYYYMM
func (this *ServerDailyStatDAO) SumUserMonthlyPeek(tx *dbs.Tx, userId int64, regionId int64, month string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	max, err := query.Between("day", month+"01", month+"32").
		Attr("userId", userId).
		Max("bytes", 0)
	if err != nil {
		return 0, err
	}
	return int64(max), nil
}

// SumUserDaily 获取某天流量总和
// day 格式为YYYYMMDD
func (this *ServerDailyStatDAO) SumUserDaily(tx *dbs.Tx, userId int64, regionId int64, day string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	return query.
		Attr("day", day).
		Attr("userId", userId).
		SumInt64("bytes", 0)
}

// SumUserDailyPeek 获取某天带宽峰值
// day 格式为YYYYMMDD
func (this *ServerDailyStatDAO) SumUserDailyPeek(tx *dbs.Tx, userId int64, regionId int64, day string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	max, err := query.
		Attr("day", day).
		Attr("userId", userId).
		Max("bytes", 0)
	if err != nil {
		return 0, err
	}
	return int64(max), nil
}

// SumMinutelyStat 获取某个分钟内的流量
// minute 格式为YYYYMMDDHHMM，并且已经格式化成每5分钟一个值
func (this *ServerDailyStatDAO) SumMinutelyStat(tx *dbs.Tx, serverId int64, minute string) (stat *pb.ServerDailyStat, err error) {
	stat = &pb.ServerDailyStat{}

	if !regexp.MustCompile(`^\d{12}$`).MatchString(minute) {
		return
	}

	one, _, err := this.Query(tx).
		Result("SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Attr("serverId", serverId).
		Attr("day", minute[:8]).
		Attr("timeFrom", minute[8:]+"00").
		FindOne()
	if err != nil {
		return nil, err
	}

	if one == nil {
		return
	}

	stat.Bytes = one.GetInt64("bytes")
	stat.CachedBytes = one.GetInt64("cachedBytes")
	stat.CountRequests = one.GetInt64("countRequests")
	stat.CountCachedRequests = one.GetInt64("countCachedRequests")
	stat.CountAttackRequests = one.GetInt64("countAttackRequests")
	stat.AttackBytes = one.GetInt64("attackBytes")
	return
}

// SumHourlyStat 获取某个小时内的流量
// hour 格式为YYYYMMDDHH
func (this *ServerDailyStatDAO) SumHourlyStat(tx *dbs.Tx, serverId int64, hour string) (stat *pb.ServerDailyStat, err error) {
	stat = &pb.ServerDailyStat{}

	if !regexp.MustCompile(`^\d{10}$`).MatchString(hour) {
		return
	}

	one, _, err := this.Query(tx).
		Result("SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Attr("serverId", serverId).
		Attr("day", hour[:8]).
		Gte("timeFrom", hour[8:]+"0000").
		Lte("timeTo", hour[8:]+"5959").
		FindOne()
	if err != nil {
		return nil, err
	}

	if one == nil {
		return
	}

	stat.Bytes = one.GetInt64("bytes")
	stat.CachedBytes = one.GetInt64("cachedBytes")
	stat.CountRequests = one.GetInt64("countRequests")
	stat.CountCachedRequests = one.GetInt64("countCachedRequests")
	stat.CountAttackRequests = one.GetInt64("countAttackRequests")
	stat.AttackBytes = one.GetInt64("attackBytes")
	return
}

// SumDailyStat 获取某天内的流量
// day 格式为YYYYMMDD
func (this *ServerDailyStatDAO) SumDailyStat(tx *dbs.Tx, serverId int64, day string) (stat *pb.ServerDailyStat, err error) {
	stat = &pb.ServerDailyStat{}

	if !regexp.MustCompile(`^\d{8}$`).MatchString(day) {
		return nil, errors.New("invalid day '" + day + "'")
	}

	one, _, err := this.Query(tx).
		Result("SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Attr("serverId", serverId).
		Attr("day", day).
		FindOne()
	if err != nil {
		return nil, err
	}

	if one == nil {
		return
	}

	stat.Bytes = one.GetInt64("bytes")
	stat.CachedBytes = one.GetInt64("cachedBytes")
	stat.CountRequests = one.GetInt64("countRequests")
	stat.CountCachedRequests = one.GetInt64("countCachedRequests")
	stat.CountAttackRequests = one.GetInt64("countAttackRequests")
	stat.AttackBytes = one.GetInt64("attackBytes")
	return
}

// SumMonthlyStat 获取某月内的流量
// month 格式为YYYYMM
func (this *ServerDailyStatDAO) SumMonthlyStat(tx *dbs.Tx, serverId int64, month string) (stat *pb.ServerDailyStat, err error) {
	stat = &pb.ServerDailyStat{}

	if !regexp.MustCompile(`^\d{6}$`).MatchString(month) {
		return
	}

	one, _, err := this.Query(tx).
		Result("SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Attr("serverId", serverId).
		Between("day", month+"01", month+"31").
		FindOne()
	if err != nil {
		return nil, err
	}

	if one == nil {
		return
	}

	stat.Bytes = one.GetInt64("bytes")
	stat.CachedBytes = one.GetInt64("cachedBytes")
	stat.CountRequests = one.GetInt64("countRequests")
	stat.CountCachedRequests = one.GetInt64("countCachedRequests")
	stat.CountAttackRequests = one.GetInt64("countAttackRequests")
	stat.AttackBytes = one.GetInt64("attackBytes")
	return
}

// FindDailyStats 按天统计
func (this *ServerDailyStatDAO) FindDailyStats(tx *dbs.Tx, serverId int64, dayFrom string, dayTo string) (result []*ServerDailyStat, err error) {
	ones, err := this.Query(tx).
		Result("SUM(bytes) AS bytes", "SUM(cachedBytes) AS cachedBytes", "SUM(countRequests) AS countRequests", "SUM(countCachedRequests) AS countCachedRequests", "SUM(countAttackRequests) AS countAttackRequests", "SUM(attackBytes) AS attackBytes", "day").
		Attr("serverId", serverId).
		Between("day", dayFrom, dayTo).
		Group("day").
		FindAll()
	if err != nil {
		return nil, err
	}

	dayMap := map[string]*ServerDailyStat{} // day => Stat
	for _, one := range ones {
		stat := one.(*ServerDailyStat)
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
			result = append(result, &ServerDailyStat{Day: day})
		}
	}

	return
}

// FindMonthlyStatsWithPlan 查找某月有套餐的流量
// month YYYYMM
func (this *ServerDailyStatDAO) FindMonthlyStatsWithPlan(tx *dbs.Tx, month string) (result []*ServerDailyStat, err error) {
	_, err = this.Query(tx).
		Between("day", month+"01", month+"32").
		Gt("planId", 0).
		Slice(&result).
		FindAll()
	return
}

// FindHourlyStats 按小时统计
func (this *ServerDailyStatDAO) FindHourlyStats(tx *dbs.Tx, serverId int64, hourFrom string, hourTo string) (result []*ServerDailyStat, err error) {
	ones, err := this.Query(tx).
		Result("SUM(bytes) AS bytes", "SUM(cachedBytes) AS cachedBytes", "SUM(countRequests) AS countRequests", "SUM(countCachedRequests) AS countCachedRequests", "SUM(countAttackRequests) AS countAttackRequests", "SUM(attackBytes) AS attackBytes", "hour").
		Attr("serverId", serverId).
		Between("hour", hourFrom, hourTo).
		Group("hour").
		FindAll()
	if err != nil {
		return nil, err
	}

	hourMap := map[string]*ServerDailyStat{} // hour => Stat
	for _, one := range ones {
		stat := one.(*ServerDailyStat)
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
			result = append(result, &ServerDailyStat{Hour: hour})
		}
	}

	return
}

// FindTopUserStats 流量排行
func (this *ServerDailyStatDAO) FindTopUserStats(tx *dbs.Tx, hourFrom string, hourTo string) (result []*ServerDailyStat, err error) {
	_, err = this.Query(tx).
		Result("userId", "SUM(bytes) AS bytes", "SUM(countRequests) AS countRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Between("hour", hourFrom, hourTo).
		Where("userId>0").
		Group("userId").
		Slice(&result).
		FindAll()
	return
}

// UpdateStatFee 设置费用
func (this *ServerDailyStatDAO) UpdateStatFee(tx *dbs.Tx, statId int64, fee float32) error {
	return this.Query(tx).
		Pk(statId).
		Set("fee", fee).
		UpdateQuickly()
}

// Clean 清理历史数据
func (this *ServerDailyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("day", day).
		Delete()
	return err
}
