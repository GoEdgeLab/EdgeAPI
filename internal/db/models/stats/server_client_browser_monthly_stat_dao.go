package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type ServerClientBrowserMonthlyStatDAO dbs.DAO

func NewServerClientBrowserMonthlyStatDAO() *ServerClientBrowserMonthlyStatDAO {
	return dbs.NewDAO(&ServerClientBrowserMonthlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerClientBrowserMonthlyStats",
			Model:  new(ServerClientBrowserMonthlyStat),
			PkName: "id",
		},
	}).(*ServerClientBrowserMonthlyStatDAO)
}

var SharedServerClientBrowserMonthlyStatDAO *ServerClientBrowserMonthlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerClientBrowserMonthlyStatDAO = NewServerClientBrowserMonthlyStatDAO()
	})
}

// 增加数量
func (this *ServerClientBrowserMonthlyStatDAO) IncreaseMonthlyCount(tx *dbs.Tx, serverId int64, browserId int64, version string, month string, count int64) error {
	if len(month) != 6 {
		return errors.New("invalid month '" + month + "'")
	}
	err := this.Query(tx).
		Param("count", count).
		InsertOrUpdateQuickly(maps.Map{
			"serverId":  serverId,
			"browserId": browserId,
			"version":   version,
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
func (this *ServerClientBrowserMonthlyStatDAO) ListStats(tx *dbs.Tx, serverId int64, month string, offset int64, size int64) (result []*ServerClientBrowserMonthlyStat, err error) {
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
