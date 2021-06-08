package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"regexp"
)

type ServerDailyStatDAO dbs.DAO

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
	for _, stat := range stats {
		day := timeutil.FormatTime("Ymd", stat.CreatedAt)
		timeFrom := timeutil.FormatTime("His", stat.CreatedAt)
		timeTo := timeutil.FormatTime("His", stat.CreatedAt+5*60-1) // 5分钟

		_, _, err := this.Query(tx).
			Param("bytes", stat.Bytes).
			Param("cachedBytes", stat.CachedBytes).
			Param("countRequests", stat.CountRequests).
			Param("countCachedRequests", stat.CountCachedRequests).
			InsertOrUpdate(maps.Map{
				"serverId":            stat.ServerId,
				"regionId":            stat.RegionId,
				"bytes":               dbs.SQL("bytes+:bytes"),
				"cachedBytes":         dbs.SQL("cachedBytes+:cachedBytes"),
				"countRequests":       dbs.SQL("countRequests+:countRequests"),
				"countCachedRequests": dbs.SQL("countCachedRequests+:countCachedRequests"),
				"day":                 day,
				"timeFrom":            timeFrom,
				"timeTo":              timeTo,
			}, maps.Map{
				"bytes":               dbs.SQL("bytes+:bytes"),
				"cachedBytes":         dbs.SQL("cachedBytes+:cachedBytes"),
				"countRequests":       dbs.SQL("countRequests+:countRequests"),
				"countCachedRequests": dbs.SQL("countCachedRequests+:countCachedRequests"),
			})
		if err != nil {
			return err
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
		Where("serverId IN (SELECT id FROM "+SharedServerDAO.Table+" WHERE userId=:userId)").
		Param("userId", userId).
		SumInt64("bytes", 0)
}

// SumUserMonthlyPeek 获取某月带宽峰值
// month 格式为YYYYMM
func (this *ServerDailyStatDAO) SumUserMonthlyPeek(tx *dbs.Tx, userId int64, regionId int64, month string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	max, err := query.Between("day", month+"01", month+"32").
		Where("serverId IN (SELECT id FROM "+SharedServerDAO.Table+" WHERE userId=:userId)").
		Param("userId", userId).
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
		Where("serverId IN (SELECT id FROM "+SharedServerDAO.Table+" WHERE userId=:userId)").
		Param("userId", userId).
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
		Where("serverId IN (SELECT id FROM "+SharedServerDAO.Table+" WHERE userId=:userId)").
		Param("userId", userId).
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
		Result("SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests").
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
		Result("SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests").
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
	return
}

// SumDailyStat 获取某天内的流量
// day 格式为YYYYMMDD
func (this *ServerDailyStatDAO) SumDailyStat(tx *dbs.Tx, serverId int64, day string) (stat *pb.ServerDailyStat, err error) {
	stat = &pb.ServerDailyStat{}

	if !regexp.MustCompile(`^\d{8}$`).MatchString(day) {
		return
	}

	one, _, err := this.Query(tx).
		Result("SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests").
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
	return
}
