package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
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

// 提交数据
func (this *ServerDailyStatDAO) SaveStats(stats []*pb.ServerDailyStat) error {
	for _, stat := range stats {
		day := timeutil.FormatTime("Ymd", stat.CreatedAt)
		timeFrom := timeutil.FormatTime("His", stat.CreatedAt)
		timeTo := timeutil.FormatTime("His", stat.CreatedAt+5*60) // 5分钟

		_, _, err := this.Query().
			Param("bytes", stat.Bytes).
			InsertOrUpdate(maps.Map{
				"serverId": stat.ServerId,
				"regionId": stat.RegionId,
				"bytes":    dbs.SQL("bytes+:bytes"),
				"day":      day,
				"timeFrom": timeFrom,
				"timeTo":   timeTo,
			}, maps.Map{
				"bytes": dbs.SQL("bytes+:bytes"),
			})
		if err != nil {
			return err
		}
	}
	return nil
}
