package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

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

// 增加数量
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

// 查找单页数据
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
