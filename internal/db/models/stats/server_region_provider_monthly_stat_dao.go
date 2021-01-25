package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type ServerRegionProviderMonthlyStatDAO dbs.DAO

func NewServerRegionProviderMonthlyStatDAO() *ServerRegionProviderMonthlyStatDAO {
	return dbs.NewDAO(&ServerRegionProviderMonthlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerRegionProviderMonthlyStats",
			Model:  new(ServerRegionProviderMonthlyStat),
			PkName: "id",
		},
	}).(*ServerRegionProviderMonthlyStatDAO)
}

var SharedServerRegionProviderMonthlyStatDAO *ServerRegionProviderMonthlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerRegionProviderMonthlyStatDAO = NewServerRegionProviderMonthlyStatDAO()
	})
}

// 增加数量
func (this *ServerRegionProviderMonthlyStatDAO) IncreaseMonthlyCount(tx *dbs.Tx, serverId int64, providerId int64, month string, count int64) error {
	if len(month) != 6 {
		return errors.New("invalid month '" + month + "'")
	}
	err := this.Query(tx).
		Param("count", count).
		InsertOrUpdateQuickly(maps.Map{
			"serverId":   serverId,
			"providerId": providerId,
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
func (this *ServerRegionProviderMonthlyStatDAO) ListStats(tx *dbs.Tx, serverId int64, month string, offset int64, size int64) (result []*ServerRegionProviderMonthlyStat, err error) {
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
