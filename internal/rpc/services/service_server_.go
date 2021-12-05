package services

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strings"
	"sync"
	"time"
)

type TrafficStat struct {
	Bytes                int64
	CachedBytes          int64
	CountRequests        int64
	CountCachedRequests  int64
	CountAttackRequests  int64
	AttackBytes          int64
	PlanId               int64
	CheckingTrafficLimit bool
}

// HTTP请求统计缓存队列
var serverHTTPCountryStatMap = map[string]*TrafficStat{}    // serverId@countryId@day => *TrafficStat
var serverHTTPProvinceStatMap = map[string]int64{}          // serverId@provinceId@month => count
var serverHTTPCityStatMap = map[string]int64{}              // serverId@cityId@month => count
var serverHTTPProviderStatMap = map[string]int64{}          // serverId@providerId@month => count
var serverHTTPSystemStatMap = map[string]int64{}            // serverId@systemId@version@month => count
var serverHTTPBrowserStatMap = map[string]int64{}           // serverId@browserId@version@month => count
var serverHTTPFirewallRuleGroupStatMap = map[string]int64{} // serverId@firewallRuleGroupId@action@day => count
var serverStatLocker = sync.Mutex{}

func init() {
	var service = new(ServerService)

	dbs.OnReadyDone(func() {
		// 导入统计数据
		go func() {
			var duration = 30 * time.Minute
			if Tea.IsTesting() {
				// 测试条件下缩短时间，以便进行观察
				duration = 10 * time.Second
			}
			ticker := time.NewTicker(duration)
			for range ticker.C {
				err := service.dumpServerHTTPStats()
				if err != nil {
					remotelogs.Error("SERVER_SERVICE", err.Error())
				}
			}
		}()
	})
}

func (this *ServerService) dumpServerHTTPStats() error {
	// 地区
	{
		serverStatLocker.Lock()
		m := serverHTTPCountryStatMap
		serverHTTPCountryStatMap = map[string]*TrafficStat{}
		serverStatLocker.Unlock()
		for k, stat := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 3 {
				continue
			}

			// Monthly
			var day = pieces[2]
			if len(day) != 8 {
				return errors.New("invalid day '" + day + "'")
			}
			err := stats.SharedServerRegionCountryMonthlyStatDAO.IncreaseMonthlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), day[:6], stat.CountRequests)
			if err != nil {
				return err
			}

			// Daily
			if teaconst.IsPlus { // 非商业版暂时不记录
				err = stats.SharedServerRegionCountryDailyStatDAO.IncreaseDailyStat(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), day, stat.Bytes, stat.CountRequests, stat.AttackBytes, stat.CountAttackRequests)
				if err != nil {
					return err
				}
			}
		}
	}

	// 省份
	{
		serverStatLocker.Lock()
		m := serverHTTPProvinceStatMap
		serverHTTPProvinceStatMap = map[string]int64{}
		serverStatLocker.Unlock()
		for k, count := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 3 {
				continue
			}
			err := stats.SharedServerRegionProvinceMonthlyStatDAO.IncreaseMonthlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], count)
			if err != nil {
				return err
			}
		}
	}

	// 城市
	{
		serverStatLocker.Lock()
		m := serverHTTPCityStatMap
		serverHTTPCityStatMap = map[string]int64{}
		serverStatLocker.Unlock()
		for k, countRequests := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 3 {
				continue
			}
			err := stats.SharedServerRegionCityMonthlyStatDAO.IncreaseMonthlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], countRequests)
			if err != nil {
				return err
			}
		}
	}

	// 运营商
	{
		serverStatLocker.Lock()
		m := serverHTTPProviderStatMap
		serverHTTPProviderStatMap = map[string]int64{}
		serverStatLocker.Unlock()
		for k, count := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 3 {
				continue
			}
			err := stats.SharedServerRegionProviderMonthlyStatDAO.IncreaseMonthlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], count)
			if err != nil {
				return err
			}
		}
	}

	// 操作系统
	{
		serverStatLocker.Lock()
		m := serverHTTPSystemStatMap
		serverHTTPSystemStatMap = map[string]int64{}
		serverStatLocker.Unlock()
		for k, count := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 4 {
				continue
			}
			err := stats.SharedServerClientSystemMonthlyStatDAO.IncreaseMonthlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], pieces[3], count)
			if err != nil {
				return err
			}
		}
	}

	// 浏览器
	{
		serverStatLocker.Lock()
		m := serverHTTPBrowserStatMap
		serverHTTPBrowserStatMap = map[string]int64{}
		serverStatLocker.Unlock()
		for k, count := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 4 {
				continue
			}
			err := stats.SharedServerClientBrowserMonthlyStatDAO.IncreaseMonthlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], pieces[3], count)
			if err != nil {
				return err
			}
		}
	}

	// 防火墙
	{
		serverStatLocker.Lock()
		m := serverHTTPFirewallRuleGroupStatMap
		serverHTTPFirewallRuleGroupStatMap = map[string]int64{}
		serverStatLocker.Unlock()
		for k, count := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 4 {
				continue
			}

			// 按天统计
			err := stats.SharedServerHTTPFirewallDailyStatDAO.IncreaseDailyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], pieces[3], count)
			if err != nil {
				return err
			}

			// 按小时统计
			err = stats.SharedServerHTTPFirewallHourlyStatDAO.IncreaseHourlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], pieces[3]+timeutil.Format("H"), count)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
