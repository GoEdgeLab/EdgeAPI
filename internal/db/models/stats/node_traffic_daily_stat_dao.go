package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type NodeTrafficDailyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(24 * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedNodeTrafficDailyStatDAO.Clean(nil, 60) // 只保留60天
				if err != nil {
					remotelogs.Error("NodeTrafficDailyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

func NewNodeTrafficDailyStatDAO() *NodeTrafficDailyStatDAO {
	return dbs.NewDAO(&NodeTrafficDailyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeTrafficDailyStats",
			Model:  new(NodeTrafficDailyStat),
			PkName: "id",
		},
	}).(*NodeTrafficDailyStatDAO)
}

var SharedNodeTrafficDailyStatDAO *NodeTrafficDailyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeTrafficDailyStatDAO = NewNodeTrafficDailyStatDAO()
	})
}

// IncreaseDailyStat 增加统计数据
func (this *NodeTrafficDailyStatDAO) IncreaseDailyStat(tx *dbs.Tx, clusterId int64, role string, nodeId int64, day string, bytes int64, cachedBytes int64, countRequests int64, countCachedRequests int64) error {
	if len(day) != 8 {
		return errors.New("invalid day '" + day + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		Param("cachedBytes", cachedBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		InsertOrUpdateQuickly(maps.Map{
			"clusterId":           clusterId,
			"role":                role,
			"nodeId":              nodeId,
			"day":                 day,
			"bytes":               bytes,
			"cachedBytes":         cachedBytes,
			"countRequests":       countRequests,
			"countCachedRequests": countCachedRequests,
		}, maps.Map{
			"bytes":               dbs.SQL("bytes+:bytes"),
			"cachedBytes":         dbs.SQL("cachedBytes+:cachedBytes"),
			"countRequests":       dbs.SQL("countRequests+:countRequests"),
			"countCachedRequests": dbs.SQL("countCachedRequests+:countCachedRequests"),
		})
	if err != nil {
		return err
	}
	return nil
}

// Clean 清理历史数据
func (this *NodeTrafficDailyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("day", day).
		Delete()
	return err
}
