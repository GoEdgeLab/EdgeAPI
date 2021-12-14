package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedServerRegionCountryMonthlyStatDAO.Clean(nil)
				if err != nil {
					remotelogs.Error("SharedServerRegionCountryMonthlyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

type ServerRegionCountryMonthlyStatDAO dbs.DAO

func NewServerRegionCountryMonthlyStatDAO() *ServerRegionCountryMonthlyStatDAO {
	return dbs.NewDAO(&ServerRegionCountryMonthlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerRegionCountryMonthlyStats",
			Model:  new(ServerRegionCountryMonthlyStat),
			PkName: "id",
		},
	}).(*ServerRegionCountryMonthlyStatDAO)
}

var SharedServerRegionCountryMonthlyStatDAO *ServerRegionCountryMonthlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerRegionCountryMonthlyStatDAO = NewServerRegionCountryMonthlyStatDAO()
	})
}

// IncreaseMonthlyCount 增加数量
func (this *ServerRegionCountryMonthlyStatDAO) IncreaseMonthlyCount(tx *dbs.Tx, serverId int64, countryId int64, month string, count int64) error {
	if len(month) != 6 {
		return errors.New("invalid month '" + month + "'")
	}
	err := this.Query(tx).
		Param("count", count).
		InsertOrUpdateQuickly(maps.Map{
			"serverId":  serverId,
			"countryId": countryId,
			"month":     month,
			"count":     count,
		}, maps.Map{
			"count": dbs.SQL("count+:count"),
		})
	if err != nil {
		return err
	}
	return nil
}

// ListStats 查找单页数据
func (this *ServerRegionCountryMonthlyStatDAO) ListStats(tx *dbs.Tx, serverId int64, month string, offset int64, size int64) (result []*ServerRegionCountryMonthlyStat, err error) {
	query := this.Query(tx).
		Attr("serverId", serverId).
		Attr("month", month).
		Offset(offset).
		Limit(size).
		Slice(&result).
		Desc("count")
	_, err = query.FindAll()
	return
}

// Clean 清理统计数据
func (this *ServerRegionCountryMonthlyStatDAO) Clean(tx *dbs.Tx) error {
	// 只保留两个月的
	var month = timeutil.Format("Ym", time.Now().AddDate(0, -2, 0))
	_, err := this.Query(tx).
		Lte("month", month).
		Delete()
	return err
}
