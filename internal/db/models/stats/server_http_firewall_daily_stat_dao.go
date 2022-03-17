package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type ServerHTTPFirewallDailyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedServerHTTPFirewallDailyStatDAO.Clean(nil, 30) // 只保留N天
				if err != nil {
					remotelogs.Error("ServerHTTPFirewallDailyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

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

// IncreaseDailyCount 增加数量
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

// SumDailyCount 计算某天的数据
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

// GroupDailyCount 查询规则分组数量
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

// FindDailyStats 查询某个日期段内的记录
func (this *ServerHTTPFirewallDailyStatDAO) FindDailyStats(tx *dbs.Tx, userId int64, serverId int64, actions []string, dayFrom string, dayTo string) (result []*ServerHTTPFirewallDailyStat, err error) {
	if len(actions) == 0 {
		return nil, nil
	}
	query := this.Query(tx).
		Between("day", dayFrom, dayTo).
		Attr("action", actions)
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

// Clean 清理历史数据
func (this *ServerHTTPFirewallDailyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("day", day).
		Delete()
	return err
}
