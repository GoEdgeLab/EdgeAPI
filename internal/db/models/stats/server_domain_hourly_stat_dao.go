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
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"sort"
	"strings"
	"sync"
	"time"
)

type ServerDomainHourlyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		goman.New(func() {
			for range ticker.C {
				err := SharedServerDomainHourlyStatDAO.CleanDefaultDays(nil, 7) // 只保留 N 天
				if err != nil {
					remotelogs.Error("ServerDomainHourlyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		})
	})
}

func NewServerDomainHourlyStatDAO() *ServerDomainHourlyStatDAO {
	return dbs.NewDAO(&ServerDomainHourlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerDomainHourlyStats",
			Model:  new(ServerDomainHourlyStat),
			PkName: "id",
		},
	}).(*ServerDomainHourlyStatDAO)
}

var SharedServerDomainHourlyStatDAO *ServerDomainHourlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerDomainHourlyStatDAO = NewServerDomainHourlyStatDAO()
	})
}

// PartitionTable 获取分区表格名称
func (this *ServerDomainHourlyStatDAO) PartitionTable(domain string) string {
	if len(domain) == 0 {
		return this.Table + "_0"
	}
	if (domain[0] >= '0' && domain[0] <= '9') || (domain[0] >= 'a' && domain[0] <= 'z') || (domain[0] >= 'A' && domain[0] <= 'Z') {
		return this.Table + "_" + strings.ToLower(string(domain[0]))
	}

	return this.Table + "_0"
}

// FindAllPartitionTables 获取所有表格名称
func (this *ServerDomainHourlyStatDAO) FindAllPartitionTables() []string {
	var tables = []string{}
	for i := '0'; i <= '9'; i++ {
		tables = append(tables, this.Table+"_"+string(i))
	}
	for i := 'a'; i <= 'z'; i++ {
		tables = append(tables, this.Table+"_"+string(i))
	}
	return tables
}

// IncreaseHourlyStat 增加统计数据
func (this *ServerDomainHourlyStatDAO) IncreaseHourlyStat(tx *dbs.Tx, clusterId int64, nodeId int64, serverId int64, domain string, hour string, bytes int64, cachedBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64, attackBytes int64) error {
	if len(hour) != 10 {
		return errors.New("invalid hour '" + hour + "'")
	}
	if len(domain) == 0 || len(domain) > 64 {
		return nil
	}
	err := this.Query(tx).
		Table(this.PartitionTable(domain)).
		Param("bytes", bytes).
		Param("cachedBytes", cachedBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		Param("countAttackRequests", countAttackRequests).
		Param("attackBytes", attackBytes).
		InsertOrUpdateQuickly(maps.Map{
			"clusterId":           clusterId,
			"nodeId":              nodeId,
			"serverId":            serverId,
			"hour":                hour,
			"domain":              domain,
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

// FindTopDomainStats 取得一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStats(tx *dbs.Tx, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, resultErr error) {
	var tables = this.FindAllPartitionTables()
	var wg = sync.WaitGroup{}
	wg.Add(len(tables))
	var locker = sync.Mutex{}

	for _, table := range tables {
		go func(table string) {
			defer wg.Done()

			var topResults = []*ServerDomainHourlyStat{}

			// TODO 节点如果已经被删除，则忽略
			_, err := this.Query(tx).
				Table(table).
				Between("hour", hourFrom, hourTo).
				Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
				Group("domain").
				Desc("countRequests").
				Limit(size).
				Slice(&topResults).
				FindAll()
			if err != nil {
				resultErr = err
				return
			}

			if len(topResults) > 0 {
				locker.Lock()
				result = append(result, topResults...)
				locker.Unlock()
			}
		}(table)
	}
	wg.Wait()

	sort.Slice(result, func(i, j int) bool {
		return result[i].CountRequests > result[j].CountRequests
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	return
}

// FindTopDomainStatsWithAttack 取得一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStatsWithAttack(tx *dbs.Tx, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, resultErr error) {
	var tables = this.FindAllPartitionTables()
	var wg = sync.WaitGroup{}
	wg.Add(len(tables))
	var locker = sync.Mutex{}

	for _, table := range tables {
		go func(table string) {
			defer wg.Done()

			var topResults = []*ServerDomainHourlyStat{}

			// TODO 节点如果已经被删除，则忽略
			_, err := this.Query(tx).
				Table(table).
				Gt("countAttackRequests", 0).
				Between("hour", hourFrom, hourTo).
				Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
				Group("domain").
				Desc("countRequests").
				Limit(size).
				Slice(&topResults).
				FindAll()
			if err != nil {
				resultErr = err
				return
			}

			if len(topResults) > 0 {
				locker.Lock()
				result = append(result, topResults...)
				locker.Unlock()
			}
		}(table)
	}
	wg.Wait()

	sort.Slice(result, func(i, j int) bool {
		return result[i].CountRequests > result[j].CountRequests
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	return
}

// FindTopDomainStatsWithClusterId 取得集群上的一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStatsWithClusterId(tx *dbs.Tx, clusterId int64, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, resultErr error) {
	var tables = this.FindAllPartitionTables()
	var wg = sync.WaitGroup{}
	wg.Add(len(tables))
	var locker = sync.Mutex{}

	for _, table := range tables {
		go func(table string) {
			defer wg.Done()

			var topResults = []*ServerDomainHourlyStat{}

			// TODO 节点如果已经被删除，则忽略
			_, err := this.Query(tx).
				Table(table).
				Attr("clusterId", clusterId).
				Between("hour", hourFrom, hourTo).
				UseIndex("hour").
				Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
				Group("domain").
				Desc("countRequests").
				Limit(size).
				Slice(&topResults).
				FindAll()
			if err != nil {
				resultErr = err
				return
			}

			if len(topResults) > 0 {
				locker.Lock()
				result = append(result, topResults...)
				locker.Unlock()
			}
		}(table)
	}
	wg.Wait()

	sort.Slice(result, func(i, j int) bool {
		return result[i].CountRequests > result[j].CountRequests
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	return
}

// FindTopDomainStatsWithNodeId 取得节点上的一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStatsWithNodeId(tx *dbs.Tx, nodeId int64, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, resultErr error) {
	var tables = this.FindAllPartitionTables()
	var wg = sync.WaitGroup{}
	wg.Add(len(tables))
	var locker = sync.Mutex{}

	for _, table := range tables {
		go func(table string) {
			defer wg.Done()

			var topResults = []*ServerDomainHourlyStat{}

			// TODO 节点如果已经被删除，则忽略
			_, err := this.Query(tx).
				Table(table).
				Attr("nodeId", nodeId).
				Between("hour", hourFrom, hourTo).
				UseIndex("hour").
				Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
				Group("domain").
				Desc("countRequests").
				Limit(size).
				Slice(&topResults).
				FindAll()
			if err != nil {
				resultErr = err
				return
			}

			if len(topResults) > 0 {
				locker.Lock()
				result = append(result, topResults...)
				locker.Unlock()
			}
		}(table)
	}
	wg.Wait()

	sort.Slice(result, func(i, j int) bool {
		return result[i].CountRequests > result[j].CountRequests
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	return
}

// FindTopDomainStatsWithServerId 取得某个服务的一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStatsWithServerId(tx *dbs.Tx, serverId int64, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, resultErr error) {
	var tables = this.FindAllPartitionTables()
	var wg = sync.WaitGroup{}
	wg.Add(len(tables))
	var locker = sync.Mutex{}

	for _, table := range tables {
		go func(table string) {
			defer wg.Done()

			var topResults = []*ServerDomainHourlyStat{}

			// TODO 节点如果已经被删除，则忽略
			_, err := this.Query(tx).
				Table(table).
				Attr("serverId", serverId).
				Between("hour", hourFrom, hourTo).
				UseIndex("serverId", "hour").
				Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
				Group("domain").
				Desc("countRequests").
				Limit(size).
				Slice(&topResults).
				FindAll()
			if err != nil {
				resultErr = err
				return
			}

			if len(topResults) > 0 {
				locker.Lock()
				result = append(result, topResults...)
				locker.Unlock()
			}
		}(table)
	}
	wg.Wait()

	sort.Slice(result, func(i, j int) bool {
		return result[i].CountRequests > result[j].CountRequests
	})

	if len(result) > types.Int(size) {
		result = result[:types.Int(size)]
	}

	return
}

// CleanDays 清理历史数据
func (this *ServerDomainHourlyStatDAO) CleanDays(tx *dbs.Tx, days int) error {
	var hour = timeutil.Format("Ymd00", time.Now().AddDate(0, 0, -days))
	for _, table := range this.FindAllPartitionTables() {
		_, err := this.Query(tx).
			Table(table).
			Lt("hour", hour).
			Delete()
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *ServerDomainHourlyStatDAO) CleanDefaultDays(tx *dbs.Tx, defaultDays int) error {
	databaseConfig, err := models.SharedSysSettingDAO.ReadDatabaseConfig(tx)
	if err != nil {
		return err
	}

	if databaseConfig != nil && databaseConfig.ServerDomainHourlyStat.Clean.Days > 0 {
		defaultDays = databaseConfig.ServerDomainHourlyStat.Clean.Days
	}
	if defaultDays <= 0 {
		defaultDays = 7
	}

	return this.CleanDays(tx, defaultDays)
}
