package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type NodeTrafficDailyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedNodeTrafficDailyStatDAO.Clean(nil, 30) // 只保留N天
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
func (this *NodeTrafficDailyStatDAO) IncreaseDailyStat(tx *dbs.Tx, clusterId int64, role string, nodeId int64, day string, bytes int64, cachedBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64, attackBytes int64) error {
	if len(day) != 8 {
		return errors.New("invalid day '" + day + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		Param("cachedBytes", cachedBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		Param("countAttackRequests", countAttackRequests).
		Param("attackBytes", attackBytes).
		InsertOrUpdateQuickly(maps.Map{
			"clusterId":           clusterId,
			"role":                role,
			"nodeId":              nodeId,
			"day":                 day,
			"bytes":               bytes,
			"cachedBytes":         cachedBytes,
			"countRequests":       countRequests,
			"countCachedRequests": countCachedRequests,
			"countAttackRequests": countAttackRequests,
			"attackBytes":         attackBytes,
		}, maps.Map{
			"bytes":               dbs.SQL("bytes+:bytes"),
			"cachedBytes":         dbs.SQL("cachedBytes+:cachedBytes"),
			"countRequests":       dbs.SQL("countRequests+:countRequests"),
			"countCachedRequests": dbs.SQL("countCachedRequests+:countCachedRequests"),
			"countAttackRequests": dbs.SQL("countAttackRequests+:countAttackRequests"),
			"attackBytes":         dbs.SQL("attackBytes+:attackBytes"),
		})
	if err != nil {
		return err
	}
	return nil
}

// FindDailyStats 获取日期之间统计
func (this *NodeTrafficDailyStatDAO) FindDailyStats(tx *dbs.Tx, role string, nodeId int64, dayFrom string, dayTo string) (result []*NodeTrafficDailyStat, err error) {
	ones, err := this.Query(tx).
		Attr("nodeId", nodeId).
		Attr("role", role).
		Between("day", dayFrom, dayTo).
		FindAll()
	if err != nil {
		return nil, err
	}
	dayMap := map[string]*NodeTrafficDailyStat{} // day => Stat
	for _, one := range ones {
		stat := one.(*NodeTrafficDailyStat)
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
			result = append(result, &NodeTrafficDailyStat{Day: day})
		}
	}
	return result, nil
}

// Clean 清理历史数据
func (this *NodeTrafficDailyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("day", day).
		Delete()
	return err
}
