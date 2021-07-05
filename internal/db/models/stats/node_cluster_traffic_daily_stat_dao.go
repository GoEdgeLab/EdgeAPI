package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type NodeClusterTrafficDailyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(24 * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedNodeClusterTrafficDailyStatDAO.Clean(nil, 60) // 只保留60天
				if err != nil {
					remotelogs.Error("NodeClusterTrafficDailyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

func NewNodeClusterTrafficDailyStatDAO() *NodeClusterTrafficDailyStatDAO {
	return dbs.NewDAO(&NodeClusterTrafficDailyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeClusterTrafficDailyStats",
			Model:  new(NodeClusterTrafficDailyStat),
			PkName: "id",
		},
	}).(*NodeClusterTrafficDailyStatDAO)
}

var SharedNodeClusterTrafficDailyStatDAO *NodeClusterTrafficDailyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeClusterTrafficDailyStatDAO = NewNodeClusterTrafficDailyStatDAO()
	})
}

// IncreaseDailyStat 增加统计数据
func (this *NodeClusterTrafficDailyStatDAO) IncreaseDailyStat(tx *dbs.Tx, clusterId int64, day string, bytes int64, cachedBytes int64, countRequests int64, countCachedRequests int64) error {
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

// FindDailyStats 获取日期之间统计
func (this *NodeClusterTrafficDailyStatDAO) FindDailyStats(tx *dbs.Tx, clusterId int64, dayFrom string, dayTo string) (result []*NodeClusterTrafficDailyStat, err error) {
	ones, err := this.Query(tx).
		Attr("clusterId", clusterId).
		Between("day", dayFrom, dayTo).
		FindAll()
	if err != nil {
		return nil, err
	}
	dayMap := map[string]*NodeClusterTrafficDailyStat{} // day => Stat
	for _, one := range ones {
		stat := one.(*NodeClusterTrafficDailyStat)
		dayMap[stat.Day] = stat
	}
	days, err := utils.RangeDays(dayFrom, dayTo)
	if err != nil {
		return nil, err
	}
	for _, day := range days {
		stat, ok := dayMap[day]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &NodeClusterTrafficDailyStat{Day: day})
		}
	}
	return result, nil
}

// Clean 清理历史数据
func (this *NodeClusterTrafficDailyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("day", day).
		Delete()
	return err
}
