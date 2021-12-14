package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
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
				err := SharedServerRegionCityMonthlyStatDAO.Clean(nil)
				if err != nil {
					remotelogs.Error("SharedServerRegionCityMonthlyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

type ServerRegionCityMonthlyStatDAO dbs.DAO

func NewServerRegionCityMonthlyStatDAO() *ServerRegionCityMonthlyStatDAO {
	return dbs.NewDAO(&ServerRegionCityMonthlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerRegionCityMonthlyStats",
			Model:  new(ServerRegionCityMonthlyStat),
			PkName: "id",
		},
	}).(*ServerRegionCityMonthlyStatDAO)
}

var SharedServerRegionCityMonthlyStatDAO *ServerRegionCityMonthlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerRegionCityMonthlyStatDAO = NewServerRegionCityMonthlyStatDAO()
	})
}

// IncreaseMonthlyCount 增加数量
func (this *ServerRegionCityMonthlyStatDAO) IncreaseMonthlyCount(tx *dbs.Tx, serverId int64, cityId int64, month string, count int64) error {
	if len(month) != 6 {
		return errors.New("invalid month '" + month + "'")
	}
	err := this.Query(tx).
		Param("count", count).
		InsertOrUpdateQuickly(maps.Map{
			"serverId": serverId,
			"cityId":   cityId,
			"month":    month,
			"count":    count,
		}, maps.Map{
			"count": dbs.SQL("count+:count"),
		})
	if err != nil {
		return err
	}
	return nil
}

// ListStats 查找单页数据
func (this *ServerRegionCityMonthlyStatDAO) ListStats(tx *dbs.Tx, serverId int64, month string, countryId int64, provinceId int64, offset int64, size int64) (result []*ServerRegionCityMonthlyStat, err error) {
	query := this.Query(tx).
		Attr("serverId", serverId).
		Attr("month", month).
		Offset(offset).
		Limit(size).
		Slice(&result).
		Desc("count")
	if countryId > 0 {
		query.Where("cityId IN (SELECT id FROM "+regions.SharedRegionCityDAO.Table+" WHERE provinceId IN (SELECT id FROM "+regions.SharedRegionProvinceDAO.Table+" WHERE countryId=:countryId AND state=1) AND state=1)").
			Param("countryId", countryId)
	}
	if provinceId > 0 {
		query.Where("cityId IN (SELECT id FROM "+regions.SharedRegionCityDAO.Table+" WHERE provinceId=:provinceId AND state=1)").
			Param("provinceId", provinceId)
	}
	_, err = query.FindAll()
	return
}

// Clean 清理统计数据
func (this *ServerRegionCityMonthlyStatDAO) Clean(tx *dbs.Tx) error {
	// 只保留两个月的
	var month = timeutil.Format("Ym", time.Now().AddDate(0, -2, 0))
	_, err := this.Query(tx).
		Lte("month", month).
		Delete()
	return err
}
