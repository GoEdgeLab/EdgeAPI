package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
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

type ServerBandwidthStatDAO dbs.DAO

const (
	ServerBandwidthStatTablePartials = 20 // 分表数量
)

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedServerBandwidthStatDAO.Clean(nil)
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
func (this *ServerBandwidthStatDAO) UpdateServerBandwidth(tx *dbs.Tx, userId int64, serverId int64, day string, timeAt string, bytes int64) error {
	if serverId <= 0 {
		return errors.New("invalid server id '" + types.String(serverId) + "'")
	}

	return this.Query(tx).
		Table(this.partialTable(serverId)).
		Param("bytes", bytes).
		InsertOrUpdateQuickly(maps.Map{
			"userId":   userId,
			"serverId": serverId,
			"day":      day,
			"timeAt":   timeAt,
			"bytes":    bytes,
		}, maps.Map{
			"bytes": dbs.SQL("bytes+:bytes"),
		})
}

// FindMinutelyPeekBandwidthBytes 获取某分钟的带宽峰值
// day YYYYMMDD
// minute HHII
func (this *ServerBandwidthStatDAO) FindMinutelyPeekBandwidthBytes(tx *dbs.Tx, serverId int64, day string, minute string) (int64, error) {
	return this.Query(tx).
		Table(this.partialTable(serverId)).
		Result("bytes").
		Attr("serverId", serverId).
		Attr("day", day).
		Attr("timeAt", minute).
		FindInt64Col(0)
}

// FindDailyPeekBandwidthBytes 获取某天的带宽峰值
// day YYYYMMDD
func (this *ServerBandwidthStatDAO) FindDailyPeekBandwidthBytes(tx *dbs.Tx, serverId int64, day string) (int64, error) {
	return this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("day", day).
		Result("MAX(bytes)").
		FindInt64Col(0)
}

// FindMonthlyPeekBandwidthBytes 获取某月的带宽峰值
// month YYYYMM
func (this *ServerBandwidthStatDAO) FindMonthlyPeekBandwidthBytes(tx *dbs.Tx, serverId int64, month string) (int64, error) {
	return this.Query(tx).
		Table(this.partialTable(serverId)).
		Between("day", month+"01", month+"31").
		Result("MAX(bytes)").
		FindInt64Col(0)
}

// FindServerStats 查找某个时间段的带宽统计
// 参数：
//   - day YYYYMMDD
//   - timeAt HHII
func (this *ServerBandwidthStatDAO) FindServerStats(tx *dbs.Tx, serverId int64, day string, timeFrom string, timeTo string) (result []*ServerBandwidthStat, err error) {
	_, err = this.Query(tx).
		Table(this.partialTable(serverId)).
		Attr("serverId", serverId).
		Attr("day", day).
		Between("timeAt", timeFrom, timeTo).
		Slice(&result).
		FindAll()
	return
}

// FindMonthlyPercentile 获取某月内百分位
func (this *ServerBandwidthStatDAO) FindMonthlyPercentile(tx *dbs.Tx, serverId int64, month string, percentile int) (result int64, err error) {
	if percentile <= 0 {
		percentile = 95
	}

	// 如果是100%以上，则快速返回
	if percentile >= 100 {
		result, err = this.Query(tx).
			Table(this.partialTable(serverId)).
			Result("bytes").
			Attr("serverId", serverId).
			Between("day", month+"01", month+"31").
			Desc("bytes").
			Limit(1).
			FindInt64Col(0)
		return
	}

	// 总数量
	total, err := this.Query(tx).
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
	result, err = this.Query(tx).
		Table(this.partialTable(serverId)).
		Result("bytes").
		Attr("serverId", serverId).
		Between("day", month+"01", month+"31").
		Desc("bytes").
		Offset(offset).
		Limit(1).
		FindInt64Col(0)

	return
}

// Clean 清理过期数据
func (this *ServerBandwidthStatDAO) Clean(tx *dbs.Tx) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -62)) // 保留大约2个月的数据
	return this.runBatch(func(table string, locker *sync.Mutex) error {
		_, err := this.Query(tx).
			Table(table).
			Lt("day", day).
			Delete()
		return err
	})
}

// 批量执行
func (this *ServerBandwidthStatDAO) runBatch(f func(table string, locker *sync.Mutex) error) error {
	var locker = &sync.Mutex{}
	var wg = sync.WaitGroup{}
	wg.Add(ServerBandwidthStatTablePartials)
	var resultErr error
	for i := 0; i < ServerBandwidthStatTablePartials; i++ {
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
	return this.Table + "_" + types.String(serverId%int64(ServerBandwidthStatTablePartials))
}
