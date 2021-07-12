package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type ServerHTTPFirewallHourlyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(24 * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedServerHTTPFirewallHourlyStatDAO.Clean(nil, 60) // 只保留60天
				if err != nil {
					remotelogs.Error("ServerHTTPFirewallHourlyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

func NewServerHTTPFirewallHourlyStatDAO() *ServerHTTPFirewallHourlyStatDAO {
	return dbs.NewDAO(&ServerHTTPFirewallHourlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerHTTPFirewallHourlyStats",
			Model:  new(ServerHTTPFirewallHourlyStat),
			PkName: "id",
		},
	}).(*ServerHTTPFirewallHourlyStatDAO)
}

var SharedServerHTTPFirewallHourlyStatDAO *ServerHTTPFirewallHourlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerHTTPFirewallHourlyStatDAO = NewServerHTTPFirewallHourlyStatDAO()
	})
}

// IncreaseHourlyCount 增加数量
func (this *ServerHTTPFirewallHourlyStatDAO) IncreaseHourlyCount(tx *dbs.Tx, serverId int64, firewallRuleGroupId int64, action string, hour string, count int64) error {
	if len(hour) != 10 {
		return errors.New("invalid hour '" + hour + "'")
	}
	err := this.Query(tx).
		Param("count", count).
		InsertOrUpdateQuickly(maps.Map{
			"serverId":                serverId,
			"day":                     hour[:8],
			"hour":                    hour,
			"httpFirewallRuleGroupId": firewallRuleGroupId,
			"action":                  action,
			"count":                   count,
		}, maps.Map{
			"count": dbs.SQL("count+:count"),
		})
	if err != nil {
		return err
	}
	return nil
}

// SumHourlyCount 计算某天的数据
func (this *ServerHTTPFirewallHourlyStatDAO) SumHourlyCount(tx *dbs.Tx, userId int64, serverId int64, action string, dayFrom string, dayTo string) (int64, error) {
	query := this.Query(tx).
		Between("day", dayFrom, dayTo)
	if serverId > 0 {
		query.Attr("serverId", serverId)
	} else if userId > 0 {
		query.Where("serverId IN (SELECT id FROM "+models.SharedServerDAO.Table+" WHERE userId=:userId AND state=1)").
			Param("userId", userId)
	}
	if len(action) > 0 {
		query.Attr("action", action)
	}
	return query.SumInt64("count", 0)
}

// GroupHourlyCount 查询规则分组数量
func (this *ServerHTTPFirewallHourlyStatDAO) GroupHourlyCount(tx *dbs.Tx, userId int64, serverId int64, dayFrom string, dayTo string, offset int64, size int64) (result []*ServerHTTPFirewallHourlyStat, err error) {
	query := this.Query(tx).
		Between("day", dayFrom, dayTo)
	if serverId > 0 {
		query.Attr("serverId", serverId)
	} else if userId > 0 {
		query.Where("serverId IN (SELECT id FROM "+models.SharedServerDAO.Table+" WHERE userId=:userId AND state=1)").
			Param("userId", userId)
	}
	_, err = query.Group("httpFirewallRuleGroupId").
		Result("httpFirewallRuleGroupId, SUM(count) AS count").
		Desc("count").
		Offset(offset).
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindHourlyStats 查询某个日期段内的记录
func (this *ServerHTTPFirewallHourlyStatDAO) FindHourlyStats(tx *dbs.Tx, userId int64, serverId int64, action string, hourFrom string, hourTo string) (result []*ServerHTTPFirewallHourlyStat, err error) {
	query := this.Query(tx).
		Between("hour", hourFrom, hourTo).
		Attr("action", action)
	if serverId > 0 {
		query.Attr("serverId", serverId)
	} else if userId > 0 {
		query.Where("serverId IN (SELECT id FROM "+models.SharedServerDAO.Table+" WHERE userId=:userId AND state=1)").
			Param("userId", userId)
	}
	_, err = query.Group("hour").
		Result("hour, SUM(count) AS count").
		Slice(&result).
		FindAll()
	return
}

// Clean 清理历史数据
func (this *ServerHTTPFirewallHourlyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var hour = timeutil.Format("Ymd00", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("hour", hour).
		Delete()
	return err
}
