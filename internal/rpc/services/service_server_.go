package services

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strings"
	"sync"
	"time"
)

// HTTP请求统计缓存队列
var serverHTTPCountryStatMap = map[string]int64{}  // serverId@countryId@month => count
var serverHTTPProvinceStatMap = map[string]int64{} // serverId@provinceId@month => count
var serverHTTPCityStatMap = map[string]int64{}     // serverId@cityId@month => count
var serverHTTPProviderStatMap = map[string]int64{} // serverId@providerId@month => count
var serverHTTPSystemStatMap = map[string]int64{}   // serverId@systemId@version@month => count
var serverHTTPBrowserStatMap = map[string]int64{}  // serverId@browserId@version@month => count
var serverStatLocker = sync.Mutex{}

func init() {
	var service = new(ServerService)

	dbs.OnReadyDone(func() {
		// 导入统计数据
		go func() {
			var duration = 30 * time.Minute
			if Tea.IsTesting() {
				// 测试条件下缩短时间，以便进行观察
				duration = 1 * time.Minute
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
		serverHTTPCountryStatMap = map[string]int64{}
		serverStatLocker.Unlock()
		for k, count := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 3 {
				continue
			}
			err := stats.SharedServerRegionCountryMonthlyStatDAO.IncreaseMonthlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], count)
			if err != nil {
				return err
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
		for k, count := range m {
			pieces := strings.Split(k, "@")
			if len(pieces) != 3 {
				continue
			}
			err := stats.SharedServerRegionCityMonthlyStatDAO.IncreaseMonthlyCount(nil, types.Int64(pieces[0]), types.Int64(pieces[1]), pieces[2], count)
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

	return nil
}
