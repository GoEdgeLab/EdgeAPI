package models

import (
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/regexputils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"math"
	"sync"
	"time"
)

type UserPlanBandwidthStatDAO dbs.DAO

const (
	UserPlanBandwidthStatTablePartitions = 20 // 分表数量
)

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedUserPlanBandwidthStatDAO.CleanDefaultDays(nil, 100)
				if err != nil {
					remotelogs.Error("SharedUserPlanBandwidthStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

func NewUserPlanBandwidthStatDAO() *UserPlanBandwidthStatDAO {
	return dbs.NewDAO(&UserPlanBandwidthStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserPlanBandwidthStats",
			Model:  new(UserPlanBandwidthStat),
			PkName: "id",
		},
	}).(*UserPlanBandwidthStatDAO)
}

var SharedUserPlanBandwidthStatDAO *UserPlanBandwidthStatDAO

func init() {
	dbs.OnReady(func() {
		SharedUserPlanBandwidthStatDAO = NewUserPlanBandwidthStatDAO()
	})
}

// UpdateUserPlanBandwidth 写入数据
// 暂时不使用region区分
func (this *UserPlanBandwidthStatDAO) UpdateUserPlanBandwidth(tx *dbs.Tx, userId int64, userPlanId int64, regionId int64, day string, timeAt string, bandwidthBytes int64, totalBytes int64, cachedBytes int64, attackBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64, countWebsocketConnections int64) error {
	if userId <= 0 || userPlanId <= 0 {
		return nil
	}

	return this.Query(tx).
		Table(this.partialTable(userPlanId)).
		Param("bytes", bandwidthBytes).
		Param("totalBytes", totalBytes).
		Param("cachedBytes", cachedBytes).
		Param("attackBytes", attackBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		Param("countAttackRequests", countAttackRequests).
		Param("countWebsocketConnections", countWebsocketConnections).
		InsertOrUpdateQuickly(maps.Map{
			"userId":                    userId,
			"userPlanId":                userPlanId,
			"regionId":                  regionId,
			"day":                       day,
			"timeAt":                    timeAt,
			"bytes":                     bandwidthBytes,
			"totalBytes":                totalBytes,
			"avgBytes":                  totalBytes / 300,
			"cachedBytes":               cachedBytes,
			"attackBytes":               attackBytes,
			"countRequests":             countRequests,
			"countCachedRequests":       countCachedRequests,
			"countAttackRequests":       countAttackRequests,
			"countWebsocketConnections": countWebsocketConnections,
		}, maps.Map{
			"bytes":                     dbs.SQL("bytes+:bytes"),
			"avgBytes":                  dbs.SQL("(totalBytes+:totalBytes)/300"), // 因为生成SQL语句时会自动将avgBytes排在totalBytes之前，所以这里不用担心先后顺序的问题
			"totalBytes":                dbs.SQL("totalBytes+:totalBytes"),
			"cachedBytes":               dbs.SQL("cachedBytes+:cachedBytes"),
			"attackBytes":               dbs.SQL("attackBytes+:attackBytes"),
			"countRequests":             dbs.SQL("countRequests+:countRequests"),
			"countCachedRequests":       dbs.SQL("countCachedRequests+:countCachedRequests"),
			"countAttackRequests":       dbs.SQL("countAttackRequests+:countAttackRequests"),
			"countWebsocketConnections": dbs.SQL("countWebsocketConnections+:countWebsocketConnections"),
		})
}

// FindMonthlyPercentile 获取某月内百分位
func (this *UserPlanBandwidthStatDAO) FindMonthlyPercentile(tx *dbs.Tx, userPlanId int64, month string, percentile int, useAvg bool) (result int64, err error) {
	if percentile <= 0 {
		percentile = 95
	}

	// 如果是100%以上，则快速返回
	if percentile >= 100 {
		result, err = this.Query(tx).
			Table(this.partialTable(userPlanId)).
			Attr("userPlanId", userPlanId).
			Result(this.sumBytesField(useAvg)).
			Between("day", month+"01", month+"31").
			Group("day").
			Group("timeAt").
			Desc("bytes").
			Limit(1).
			FindInt64Col(0)
		return
	}

	// 总数量
	total, err := this.Query(tx).
		Table(this.partialTable(userPlanId)).
		Attr("userPlanId", userPlanId).
		Between("day", month+"01", month+"31").
		CountAttr("DISTINCT day, timeAt")
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
	result, err = this.Query(tx).
		Table(this.partialTable(userPlanId)).
		Attr("userPlanId", userPlanId).
		Result(this.sumBytesField(useAvg)).
		Between("day", month+"01", month+"31").
		Group("day").
		Group("timeAt").
		Desc("bytes").
		Offset(offset).
		Limit(1).
		FindInt64Col(0)

	return
}

// SumMonthlyBytes 读取单月总流量
func (this *UserPlanBandwidthStatDAO) SumMonthlyBytes(tx *dbs.Tx, userPlanId int64, month string) (int64, error) {
	if !regexputils.YYYYMM.MatchString(month) {
		return 0, errors.New("invalid ")
	}

	return this.Query(tx).
		Table(this.partialTable(userPlanId)).
		Attr("userPlanId", userPlanId).
		Between("day", month+"01", month+"31").
		SumInt64("totalBytes", 0)
}

// CleanDefaultDays 清理过期数据
func (this *UserPlanBandwidthStatDAO) CleanDefaultDays(tx *dbs.Tx, defaultDays int) error {
	databaseConfig, err := SharedSysSettingDAO.ReadDatabaseConfig(tx)
	if err != nil {
		return err
	}

	if databaseConfig != nil && databaseConfig.UserPlanBandwidthStat.Clean.Days > 0 {
		defaultDays = databaseConfig.UserPlanBandwidthStat.Clean.Days
	}
	if defaultDays <= 0 {
		defaultDays = 100
	}

	return this.CleanDays(tx, defaultDays)
}

// CleanDays 清理过期数据
func (this *UserPlanBandwidthStatDAO) CleanDays(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days)) // 保留大约3个月的数据
	return this.runBatch(func(table string, locker *sync.Mutex) error {
		_, err := this.Query(tx).
			Table(table).
			Lt("day", day).
			Delete()
		return err
	})
}

// 获取字节字段
func (this *UserPlanBandwidthStatDAO) bytesField(useAvg bool) string {
	if useAvg {
		return "avgBytes AS bytes"
	}
	return "bytes"
}

func (this *UserPlanBandwidthStatDAO) sumBytesField(useAvg bool) string {
	if useAvg {
		return "SUM(avgBytes) AS bytes"
	}
	return "SUM(bytes) AS bytes"
}

// 批量执行
func (this *UserPlanBandwidthStatDAO) runBatch(f func(table string, locker *sync.Mutex) error) error {
	var locker = &sync.Mutex{}
	var wg = sync.WaitGroup{}
	wg.Add(UserPlanBandwidthStatTablePartitions)
	var resultErr error
	for i := 0; i < UserPlanBandwidthStatTablePartitions; i++ {
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
func (this *UserPlanBandwidthStatDAO) partialTable(userPlanId int64) string {
	return this.Table + "_" + types.String(userPlanId%int64(UserPlanBandwidthStatTablePartitions))
}
