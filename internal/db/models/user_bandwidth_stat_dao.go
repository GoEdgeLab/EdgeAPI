package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"sync"
	"time"
)

type UserBandwidthStatDAO dbs.DAO

const (
	UserBandwidthStatTablePartials = 20
)

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedUserBandwidthStatDAO.Clean(nil)
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
func (this *UserBandwidthStatDAO) UpdateUserBandwidth(tx *dbs.Tx, userId int64, day string, timeAt string, bytes int64) error {
	if userId <= 0 {
		// 如果用户ID不大于0，则说明服务不属于任何用户，此时不需要处理
		return nil
	}

	return this.Query(tx).
		Table(this.partialTable(userId)).
		Param("bytes", bytes).
		InsertOrUpdateQuickly(maps.Map{
			"userId": userId,
			"day":    day,
			"timeAt": timeAt,
			"bytes":  bytes,
		}, maps.Map{
			"bytes": dbs.SQL("bytes+:bytes"),
		})
}

// FindUserPeekBandwidthInMonth 读取某月带宽峰值
// month YYYYMM
func (this *UserBandwidthStatDAO) FindUserPeekBandwidthInMonth(tx *dbs.Tx, userId int64, month string) (*UserBandwidthStat, error) {
	one, err := this.Query(tx).
		Table(this.partialTable(userId)).
		Attr("userId", userId).
		Between("day", month+"01", month+"31").
		Desc("bytes").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*UserBandwidthStat), nil
}

// FindUserPeekBandwidthInDay 读取某日带宽峰值
// day YYYYMMDD
func (this *UserBandwidthStatDAO) FindUserPeekBandwidthInDay(tx *dbs.Tx, userId int64, day string) (*UserBandwidthStat, error) {
	one, err := this.Query(tx).
		Table(this.partialTable(userId)).
		Attr("userId", userId).
		Attr("day", day).
		Desc("bytes").
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*UserBandwidthStat), nil
}

// Clean 清理过期数据
func (this *UserBandwidthStatDAO) Clean(tx *dbs.Tx) error {
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
