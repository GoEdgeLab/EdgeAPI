package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type ServerRegionProvinceMonthlyStatDAO dbs.DAO

func NewServerRegionProvinceMonthlyStatDAO() *ServerRegionProvinceMonthlyStatDAO {
	return dbs.NewDAO(&ServerRegionProvinceMonthlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerRegionProvinceMonthlyStats",
			Model:  new(ServerRegionProvinceMonthlyStat),
			PkName: "id",
		},
	}).(*ServerRegionProvinceMonthlyStatDAO)
}

var SharedServerRegionProvinceMonthlyStatDAO *ServerRegionProvinceMonthlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerRegionProvinceMonthlyStatDAO = NewServerRegionProvinceMonthlyStatDAO()
	})
}

// 增加数量
func (this *ServerRegionProvinceMonthlyStatDAO) IncreaseMonthlyCount(tx *dbs.Tx, serverId int64, provinceId int64, month string, count int64) error {
	if len(month) != 6 {
		return errors.New("invalid month '" + month + "'")
	}
	err := this.Query(tx).
		Param("count", count).
		InsertOrUpdateQuickly(maps.Map{
			"serverId":   serverId,
			"provinceId": provinceId,
			"month":      month,
			"count":      count,
		}, maps.Map{
			"count": dbs.SQL("count+:count"),
		})
	if err != nil {
		return err
	}
	return nil
}

// 查找单页数据
func (this *ServerRegionProvinceMonthlyStatDAO) ListStats(tx *dbs.Tx, serverId int64, month string, countryId int64, offset int64, size int64) (result []*ServerRegionProvinceMonthlyStat, err error) {
	query := this.Query(tx).
		Attr("serverId", serverId).
		Attr("month", month).
		Offset(offset).
		Limit(size).
		Slice(&result).
		Desc("count")
	if countryId > 0 {
		query.Where("id IN (SELECT id FROM "+regions.SharedRegionProvinceDAO.Table+" WHERE countryId=:countryId AND state=1)").
			Param("countryId", countryId)
	}

	_, err = query.FindAll()
	return
}
