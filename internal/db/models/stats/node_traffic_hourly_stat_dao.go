package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
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

type NodeTrafficHourlyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedNodeTrafficHourlyStatDAO.CleanDefaultDays(nil, 15) // 只保留N天
				if err != nil {
					remotelogs.Error("NodeTrafficHourlyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

func NewNodeTrafficHourlyStatDAO() *NodeTrafficHourlyStatDAO {
	return dbs.NewDAO(&NodeTrafficHourlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeTrafficHourlyStats",
			Model:  new(NodeTrafficHourlyStat),
			PkName: "id",
		},
	}).(*NodeTrafficHourlyStatDAO)
}

var SharedNodeTrafficHourlyStatDAO *NodeTrafficHourlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeTrafficHourlyStatDAO = NewNodeTrafficHourlyStatDAO()
	})
}

// IncreaseHourlyStat 增加统计数据
func (this *NodeTrafficHourlyStatDAO) IncreaseHourlyStat(tx *dbs.Tx, clusterId int64, role string, nodeId int64, hour string, bytes int64, cachedBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64, attackBytes int64) error {
	if len(hour) != 10 {
		return errors.New("invalid hour '" + hour + "'")
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
			"hour":                hour,
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

// FindHourlyStatsWithClusterId 获取小时之间统计
func (this *NodeTrafficHourlyStatDAO) FindHourlyStatsWithClusterId(tx *dbs.Tx, clusterId int64, hourFrom string, hourTo string) (result []*NodeTrafficHourlyStat, err error) {
	ones, err := this.Query(tx).
		Attr("clusterId", clusterId).
		Between("hour", hourFrom, hourTo).
		Result("hour, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("hour").
		FindAll()
	if err != nil {
		return nil, err
	}
	hourMap := map[string]*NodeTrafficHourlyStat{} // hour => Stat
	for _, one := range ones {
		stat := one.(*NodeTrafficHourlyStat)
		hourMap[stat.Hour] = stat
	}
	hours, err := utils.RangeHours(hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, hour := range hours {
		stat, ok := hourMap[hour]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &NodeTrafficHourlyStat{Hour: hour})
		}
	}
	return result, nil
}

// FindTopNodeStats 取得一定时间内的节点排行数据
func (this *NodeTrafficHourlyStatDAO) FindTopNodeStats(tx *dbs.Tx, role string, hourFrom string, hourTo string, size int64) (result []*NodeTrafficHourlyStat, err error) {
	// TODO 节点如果已经被删除，则忽略
	_, err = this.Query(tx).
		Attr("role", role).
		Between("hour", hourFrom, hourTo).
		Result("nodeId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("nodeId").
		Desc("countRequests").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindTopNodeStatsWithAttack 取得防火墙相关的节点排行数据
func (this *NodeTrafficHourlyStatDAO) FindTopNodeStatsWithAttack(tx *dbs.Tx, role string, hourFrom string, hourTo string, size int64) (result []*NodeTrafficHourlyStat, err error) {
	// TODO 节点如果已经被删除，则忽略
	_, err = this.Query(tx).
		Attr("role", role).
		Gt("countAttackRequests", 0).
		Between("hour", hourFrom, hourTo).
		Result("nodeId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("nodeId").
		Desc("countRequests").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindTopNodeStatsWithClusterId 取得集群一定时间内的节点排行数据
func (this *NodeTrafficHourlyStatDAO) FindTopNodeStatsWithClusterId(tx *dbs.Tx, role string, clusterId int64, hourFrom string, hourTo string, size int64) (result []*NodeTrafficHourlyStat, err error) {
	// TODO 节点如果已经被删除，则忽略
	_, err = this.Query(tx).
		Attr("role", role).
		Attr("clusterId", clusterId).
		Between("hour", hourFrom, hourTo).
		Result("nodeId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("nodeId").
		Desc("countRequests").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindHourlyStatsWithNodeId 获取节点小时之间统计
func (this *NodeTrafficHourlyStatDAO) FindHourlyStatsWithNodeId(tx *dbs.Tx, role string, nodeId int64, hourFrom string, hourTo string) (result []*NodeTrafficHourlyStat, err error) {
	ones, err := this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Between("hour", hourFrom, hourTo).
		Result("hour, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("hour").
		FindAll()
	if err != nil {
		return nil, err
	}
	hourMap := map[string]*NodeTrafficHourlyStat{} // hour => Stat
	for _, one := range ones {
		stat := one.(*NodeTrafficHourlyStat)
		hourMap[stat.Hour] = stat
	}
	hours, err := utils.RangeHours(hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, hour := range hours {
		stat, ok := hourMap[hour]
		if ok {
			result = append(result, stat)
		} else {
			result = append(result, &NodeTrafficHourlyStat{Hour: hour})
		}
	}
	return result, nil
}

// CleanDays 清理历史数据
func (this *NodeTrafficHourlyStatDAO) CleanDays(tx *dbs.Tx, days int) error {
	var hour = timeutil.Format("Ymd00", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("hour", hour).
		Delete()
	return err
}

func (this *NodeTrafficHourlyStatDAO) CleanDefaultDays(tx *dbs.Tx, defaultDays int) error {
	databaseConfig, err := models.SharedSysSettingDAO.ReadDatabaseConfig(tx)
	if err != nil {
		return err
	}

	if databaseConfig != nil && databaseConfig.NodeTrafficHourlyStat.Clean.Days > 0 {
		defaultDays = databaseConfig.NodeTrafficHourlyStat.Clean.Days
	}
	if defaultDays <= 0 {
		defaultDays = 15
	}

	return this.CleanDays(tx, defaultDays)
}
