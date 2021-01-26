package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

type ServerHTTPFirewallDailyStatDAO dbs.DAO

func NewServerHTTPFirewallDailyStatDAO() *ServerHTTPFirewallDailyStatDAO {
	return dbs.NewDAO(&ServerHTTPFirewallDailyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerHTTPFirewallDailyStats",
			Model:  new(ServerHTTPFirewallDailyStat),
			PkName: "id",
		},
	}).(*ServerHTTPFirewallDailyStatDAO)
}

var SharedServerHTTPFirewallDailyStatDAO *ServerHTTPFirewallDailyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerHTTPFirewallDailyStatDAO = NewServerHTTPFirewallDailyStatDAO()
	})
}

// 增加数量
func (this *ServerHTTPFirewallDailyStatDAO) IncreaseDailyCount(tx *dbs.Tx, serverId int64, firewallRuleGroupId int64, action string, day string, count int64) error {
	if len(day) != 8 {
		return errors.New("invalid day '" + day + "'")
	}
	err := this.Query(tx).
		Param("count", count).
		InsertOrUpdateQuickly(maps.Map{
			"serverId":                serverId,
			"day":                     day,
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

// 计算某天的数据
func (this *ServerHTTPFirewallDailyStatDAO) SumDailyCount(tx *dbs.Tx, userId int64, serverId int64, action string, dayFrom string, dayTo string) (int64, error) {
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

// 查询规则分组数量
func (this *ServerHTTPFirewallDailyStatDAO) GroupDailyCount(tx *dbs.Tx, userId int64, serverId int64, dayFrom string, dayTo string, offset int64, size int64) (result []*ServerHTTPFirewallDailyStat, err error) {
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

// 查询某个日期段内的记录
func (this *ServerHTTPFirewallDailyStatDAO) FindDailyStats(tx *dbs.Tx, userId int64, serverId int64, action string, dayFrom string, dayTo string) (result []*ServerHTTPFirewallDailyStat, err error) {
	query := this.Query(tx).
		Between("day", dayFrom, dayTo).
		Attr("action", action)
	if serverId > 0 {
		query.Attr("serverId", serverId)
	} else if userId > 0 {
		query.Where("serverId IN (SELECT id FROM "+models.SharedServerDAO.Table+" WHERE userId=:userId AND state=1)").
			Param("userId", userId)
	}
	_, err = query.Group("day").
		Result("day, SUM(count) AS count").
		Slice(&result).
		FindAll()
	return
}
