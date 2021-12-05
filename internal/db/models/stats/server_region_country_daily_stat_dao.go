package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
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
		go func() {
			for range ticker.C {
				err := SharedServerRegionCountryDailyStatDAO.Clean(nil)
				if err != nil {
					remotelogs.Error("ServerRegionCountryDailyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

type ServerRegionCountryDailyStatDAO dbs.DAO

func NewServerRegionCountryDailyStatDAO() *ServerRegionCountryDailyStatDAO {
	return dbs.NewDAO(&ServerRegionCountryDailyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerRegionCountryDailyStats",
			Model:  new(ServerRegionCountryDailyStat),
			PkName: "id",
		},
	}).(*ServerRegionCountryDailyStatDAO)
}

var SharedServerRegionCountryDailyStatDAO *ServerRegionCountryDailyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerRegionCountryDailyStatDAO = NewServerRegionCountryDailyStatDAO()
	})
}

// IncreaseDailyStat 增加统计
func (this *ServerRegionCountryDailyStatDAO) IncreaseDailyStat(tx *dbs.Tx, serverId int64, countryId int64, day string, bytes int64, countRequests int64, attackBytes int64, countAttackRequests int64) error {
	if len(day) != 8 {
		return errors.New("invalid day '" + day + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		Param("countRequests", countRequests).
		Param("attackBytes", attackBytes).
		Param("countAttackRequests", countAttackRequests).
		InsertOrUpdateQuickly(maps.Map{
			"serverId":            serverId,
			"countryId":           countryId,
			"day":                 day,
			"bytes":               bytes,
			"attackBytes":         attackBytes,
			"countRequests":       countRequests,
			"countAttackRequests": countAttackRequests,
		}, maps.Map{
			"bytes":               dbs.SQL("bytes+:bytes"),
			"countRequests":       dbs.SQL("countRequests+:countRequests"),
			"attackBytes":         dbs.SQL("attackBytes+:attackBytes"),
			"countAttackRequests": dbs.SQL("countAttackRequests+:countAttackRequests"),
		})
	if err != nil {
		return err
	}
	return nil
}

// ListServerStats 查找单页数据
func (this *ServerRegionCountryDailyStatDAO) ListServerStats(tx *dbs.Tx, serverId int64, day string, orderField string, offset int64, size int64) (result []*ServerRegionCountryDailyStat, err error) {
	query := this.Query(tx).
		Attr("serverId", serverId).
		Attr("day", day).
		Offset(offset).
		Limit(size).
		Slice(&result)

	switch orderField {
	case "bytes":
		query.Desc("bytes")
	case "countRequests":
		query.Desc("countRequests")
	case "attackBytes":
		query.Desc("attackBytes")
		query.Gt("attackBytes", 0)
	case "countAttackRequests":
		query.Desc("countAttackRequests")
		query.Gt("countAttackRequests", 0)
	}

	_, err = query.FindAll()
	return
}

// ListSumStats 查找总体数据
func (this *ServerRegionCountryDailyStatDAO) ListSumStats(tx *dbs.Tx, day string, orderField string, offset int64, size int64) (result []*ServerRegionCountryDailyStat, err error) {
	query := this.Query(tx).
		Attr("day", day).
		Result("countryId", "SUM(bytes) AS bytes", "SUM(countRequests) AS countRequests", "SUM(attackBytes) AS attackBytes", "SUM(countAttackRequests) AS countAttackRequests").
		Group("countryId").
		Offset(offset).
		Limit(size).
		Slice(&result)

	switch orderField {
	case "bytes":
		query.Desc("bytes")
	case "countRequests":
		query.Desc("countRequests")
	case "attackBytes":
		query.Desc("attackBytes")
		query.Gt("attackBytes", 0)
	case "countAttackRequests":
		query.Desc("countAttackRequests")
		query.Gt("countAttackRequests", 0)
	}

	_, err = query.FindAll()
	return
}

// SumDailyTotalBytes 计算总流量
func (this *ServerRegionCountryDailyStatDAO) SumDailyTotalBytes(tx *dbs.Tx, day string) (int64, error) {
	return this.Query(tx).
		Attr("day", day).
		SumInt64("bytes", 0)
}

// SumDailyTotalAttackRequests 计算总攻击次数
func (this *ServerRegionCountryDailyStatDAO) SumDailyTotalAttackRequests(tx *dbs.Tx, day string) (int64, error) {
	return this.Query(tx).
		Attr("day", day).
		SumInt64("countAttackRequests", 0)
}

// Clean 清理统计数据
func (this *ServerRegionCountryDailyStatDAO) Clean(tx *dbs.Tx) error {
	// 只保留7天的
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -7))
	_, err := this.Query(tx).
		Lte("day", day).
		Delete()
	return err
}
