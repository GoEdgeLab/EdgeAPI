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
func (this *ServerDailyStatDAO) SaveStats(tx *dbs.Tx, stats []*pb.ServerDailyStat) error {
	for _, stat := range stats {
		day := timeutil.FormatTime("Ymd", stat.CreatedAt)
		timeFrom := timeutil.FormatTime("His", stat.CreatedAt)
		timeTo := timeutil.FormatTime("His", stat.CreatedAt+5*60) // 5分钟

		_, _, err := this.Query(tx).
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

// 根据用户计算某月合计
// month 格式为YYYYMM
func (this *ServerDailyStatDAO) SumUserMonthly(tx *dbs.Tx, userId int64, regionId int64, month string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	return query.Between("day", month+"01", month+"32").
		Where("serverId IN (SELECT id FROM "+SharedServerDAO.Table+" WHERE userId=:userId)").
		Param("userId", userId).
		SumInt64("bytes", 0)
}

// 获取某月带宽峰值
// month 格式为YYYYMM
func (this *ServerDailyStatDAO) SumUserMonthlyPeek(tx *dbs.Tx, userId int64, regionId int64, month string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	max, err := query.Between("day", month+"01", month+"32").
		Where("serverId IN (SELECT id FROM "+SharedServerDAO.Table+" WHERE userId=:userId)").
		Param("userId", userId).
		Max("bytes", 0)
	if err != nil {
		return 0, err
	}
	return int64(max), nil
}

// 获取某天流量总和
// day 格式为YYYYMMDD
func (this *ServerDailyStatDAO) SumUserDaily(tx *dbs.Tx, userId int64, regionId int64, day string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	return query.
		Attr("day", day).
		Where("serverId IN (SELECT id FROM "+SharedServerDAO.Table+" WHERE userId=:userId)").
		Param("userId", userId).
		SumInt64("bytes", 0)
}

// 获取某天带宽峰值
// day 格式为YYYYMMDD
func (this *ServerDailyStatDAO) SumUserDailyPeek(tx *dbs.Tx, userId int64, regionId int64, day string) (int64, error) {
	query := this.Query(tx)
	if regionId > 0 {
		query.Attr("regionId", regionId)
	}
	max, err := query.
		Attr("day", day).
		Where("serverId IN (SELECT id FROM "+SharedServerDAO.Table+" WHERE userId=:userId)").
		Param("userId", userId).
		Max("bytes", 0)
	if err != nil {
		return 0, err
	}
	return int64(max), nil
}
