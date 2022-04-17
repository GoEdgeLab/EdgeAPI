// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package models

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// 访问日志的两个表格形式
// 括号位置需要固定，会用来读取日期和分区
var accessLogTableMainReg = regexp.MustCompile(`_(\d{8})$`)
var accessLogTablePartialReg = regexp.MustCompile(`_(\d{8})_(\d{4})$`)

var SharedHTTPAccessLogManager = NewHTTPAccessLogManager()

type HTTPAccessLogManager struct {
	currentTableMapping map[string]*httpAccessLogDefinition // dsn => def

	locker sync.Mutex
}

func NewHTTPAccessLogManager() *HTTPAccessLogManager {
	return &HTTPAccessLogManager{
		currentTableMapping: map[string]*httpAccessLogDefinition{},
	}
}

// FindTableNames 读取数据库中某日所有日志表名称
func (this *HTTPAccessLogManager) FindTableNames(db *dbs.DB, day string) ([]string, error) {
	var results = []string{}

	// 需要防止用户设置了表名自动小写
	for _, prefix := range []string{"edgeHTTPAccessLogs_" + day + "%", "edgehttpaccesslogs_" + day + "%"} {
		ones, columnNames, err := db.FindPreparedOnes(`SHOW TABLES LIKE '` + prefix + `'`)
		if err != nil {
			return nil, errors.New("query table names error: " + err.Error())
		}

		var columnName = columnNames[0]

		for _, one := range ones {
			var tableName = one[columnName].(string)

			if lists.ContainsString(results, tableName) {
				continue
			}

			if accessLogTableMainReg.MatchString(tableName) || accessLogTablePartialReg.MatchString(tableName) {
				results = append(results, tableName)
			}
		}
	}

	// 排序
	// 这里不能直接使用sort.Strings()，因为表名里面可能大小写混合
	sort.Slice(results, func(i, j int) bool {
		var name1 = results[i]
		var name2 = results[j]
		if len(name1) < len(name2) {
			return true
		}
		return strings.ToLower(name1) < strings.ToLower(name2)
	})

	return results, nil
}

// FindTables 读取数据库中某日所有日志表
func (this *HTTPAccessLogManager) FindTables(db *dbs.DB, day string) ([]*httpAccessLogDefinition, error) {
	var results = []*httpAccessLogDefinition{}
	var tableNames = []string{}

	config, err := db.Config()
	if err != nil {
		return nil, err
	}
	var cachePrefix = config.Dsn

	// 需要防止用户设置了表名自动小写
	for _, prefix := range []string{"edgeHTTPAccessLogs_" + day + "%", "edgehttpaccesslogs_" + day + "%"} {
		ones, columnNames, err := db.FindPreparedOnes(`SHOW TABLES LIKE '` + prefix + `'`)
		if err != nil {
			return nil, errors.New("query table names error: " + err.Error())
		}

		var columnName = columnNames[0]

		for _, one := range ones {
			var tableName = one[columnName].(string)

			if lists.ContainsString(tableNames, tableName) {
				continue
			}

			if accessLogTableMainReg.MatchString(tableName) {
				tableNames = append(tableNames, tableName)

				// 查找已有的表格信息，避免SHOW FIELDS
				var tableDay = tableName[strings.LastIndex(tableName, "_")+1:]
				var cacheKey = this.composeTableCacheKey(cachePrefix, tableDay)
				this.locker.Lock()
				currentTableDef, ok := this.currentTableMapping[cacheKey]
				this.locker.Unlock()
				if ok {
					results = append(results, &httpAccessLogDefinition{
						Name:          tableName,
						HasRemoteAddr: currentTableDef.HasRemoteAddr,
						HasDomain:     currentTableDef.HasDomain,
						Exists:        true,
					})
				} else {
					hasRemoteAddrField, hasDomainField, err := this.checkTableFields(db, tableName)
					if err != nil {
						return nil, err
					}

					results = append(results, &httpAccessLogDefinition{
						Name:          tableName,
						HasRemoteAddr: hasRemoteAddrField,
						HasDomain:     hasDomainField,
						Exists:        true,
					})
				}
			} else if accessLogTablePartialReg.MatchString(tableName) {
				tableNames = append(tableNames, tableName)

				results = append(results, &httpAccessLogDefinition{
					Name:          tableName,
					HasRemoteAddr: true,
					HasDomain:     true,
					Exists:        true,
				})
			}
		}
	}

	// 排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

func (this *HTTPAccessLogManager) FindPartitionTable(db *dbs.DB, day string, partition int32) (*httpAccessLogDefinition, error) {
	var tableNames []string
	if partition < 0 {
		tableList, err := this.FindTables(db, day)
		if err != nil {
			return nil, err
		}

		if len(tableList) > 0 {
			return tableList[len(tableList)-1], nil
		}

		return &httpAccessLogDefinition{
			Name:          "",
			HasRemoteAddr: false,
			HasDomain:     false,
			Exists:        false,
		}, nil
	} else if partition == 0 {
		tableNames = []string{"edgeHTTPAccessLogs_" + day, "edgehttpaccesslogs_" + day}
	} else {
		tableNames = []string{"edgeHTTPAccessLogs_" + day + "_" + fmt.Sprintf("%04d", partition), "edgehttpaccesslogs_" + day + "_" + fmt.Sprintf("%04d", partition)}
	}
	for _, tableName := range tableNames {
		hasRemoteField, hasDomainField, err := this.checkTableFields(db, tableName)
		if err != nil {
			continue
		}
		return &httpAccessLogDefinition{
			Name:          tableName,
			HasRemoteAddr: hasRemoteField,
			HasDomain:     hasDomainField,
			Exists:        true,
		}, nil
	}

	return &httpAccessLogDefinition{
		Name:          "",
		HasRemoteAddr: false,
		HasDomain:     false,
		Exists:        false,
	}, nil
}

// FindLastTable 根据日期获取上一个可以使用的表名
// 表名组成
//   - PREFIX_DAY
//   - PREFIX_DAY_0001
func (this *HTTPAccessLogManager) FindLastTable(db *dbs.DB, day string, force bool) (*httpAccessLogDefinition, error) {
	this.locker.Lock()
	defer this.locker.Unlock()

	config, err := db.Config()
	if err != nil {
		return nil, err
	}
	var cachePrefix = config.Dsn
	var cacheKey = this.composeTableCacheKey(cachePrefix, day)
	def, ok := this.currentTableMapping[cacheKey]
	if ok {
		return def, nil
	}

	def, err = this.findTableWithoutCache(db, day, force)
	if err != nil {
		return nil, err
	}

	// 只有存在的表格才缓存
	if def != nil && def.Exists {
		this.currentTableMapping[cacheKey] = def

		// 清除过时缓存
		for oldCacheKey := range this.currentTableMapping {
			var dayIndex = strings.LastIndex(oldCacheKey, "_")
			if dayIndex > 0 {
				var oldPrefix = oldCacheKey[:dayIndex]
				var oldDay = oldCacheKey[dayIndex+1:]
				if oldPrefix == cachePrefix && oldDay < day {
					delete(this.currentTableMapping, oldCacheKey)
				}
			}
		}
	}
	return def, nil
}

// CreateTable 创建访问日志表格
func (this *HTTPAccessLogManager) CreateTable(db *dbs.DB, tableName string) error {
	_, err := db.Exec("CREATE TABLE `" + tableName + "` (\n  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n  `serverId` int(11) unsigned DEFAULT '0' COMMENT '服务ID',\n  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n  `status` int(3) unsigned DEFAULT '0' COMMENT '状态码',\n  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n  `content` json DEFAULT NULL COMMENT '日志内容',\n  `requestId` varchar(128) DEFAULT NULL COMMENT '请求ID',\n  `firewallPolicyId` int(11) unsigned DEFAULT '0' COMMENT 'WAF策略ID',\n  `firewallRuleGroupId` int(11) unsigned DEFAULT '0' COMMENT 'WAF分组ID',\n  `firewallRuleSetId` int(11) unsigned DEFAULT '0' COMMENT 'WAF集ID',\n  `firewallRuleId` int(11) unsigned DEFAULT '0' COMMENT 'WAF规则ID',\n  `remoteAddr` varchar(64) DEFAULT NULL COMMENT 'IP地址',\n  `domain` varchar(128) DEFAULT NULL COMMENT '域名',\n  `requestBody` mediumblob COMMENT '请求内容',\n  `responseBody` mediumblob COMMENT '响应内容',\n  PRIMARY KEY (`id`),\n  KEY `serverId` (`serverId`),\n  KEY `nodeId` (`nodeId`),\n  KEY `serverId_status` (`serverId`,`status`),\n  KEY `requestId` (`requestId`),\n  KEY `firewallPolicyId` (`firewallPolicyId`),\n  KEY `firewallRuleGroupId` (`firewallRuleGroupId`),\n  KEY `firewallRuleSetId` (`firewallRuleSetId`),\n  KEY `firewallRuleId` (`firewallRuleId`),\n  KEY `remoteAddr` (`remoteAddr`),\n  KEY `domain` (`domain`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='访问日志';")
	if err != nil {
		if CheckSQLErrCode(err, 1050) { // Error 1050: Table 'xxx' already exists
			return nil
		}

		return err
	}

	return nil
}

// ResetTable 清除某个数据库表名缓存
func (this *HTTPAccessLogManager) ResetTable(db *dbs.DB, day string) {
	this.locker.Lock()
	defer this.locker.Unlock()

	config, err := db.Config()
	if err != nil {
		return
	}
	delete(this.currentTableMapping, this.composeTableCacheKey(config.Dsn, day))
}

// TablePartition 从表名中获取分区
func (this *HTTPAccessLogManager) TablePartition(tableName string) (partition int32) {
	if accessLogTablePartialReg.MatchString(tableName) {
		return types.Int32(accessLogTablePartialReg.FindStringSubmatch(tableName)[2])
	}

	return 0
}

// FindLatestPartition 读取最后一个分区
func (this *HTTPAccessLogManager) FindLatestPartition(day string) (int32, error) {
	var dbList = AllAccessLogDBs()
	if len(dbList) == 0 {
		return 0, errors.New("no valid database")
	}

	var partitions = []int32{}
	var locker sync.Mutex

	var wg = sync.WaitGroup{}
	wg.Add(len(dbList))

	var lastErr error
	for _, db := range dbList {
		go func(db *dbs.DB) {
			defer wg.Done()

			names, err := this.FindTableNames(db, day)
			if err != nil {
				lastErr = err
			}
			for _, name := range names {
				var partition = this.TablePartition(name)
				locker.Lock()
				if !lists.Contains(partitions, partition) {
					partitions = append(partitions, partition)
				}
				locker.Unlock()
			}
		}(db)
	}
	wg.Wait()

	if lastErr != nil {
		return 0, lastErr
	}

	if len(partitions) == 0 {
		return 0, nil
	}

	return partitions[len(partitions)-1], nil
}

// 查找某个表格
func (this *HTTPAccessLogManager) findTableWithoutCache(db *dbs.DB, day string, force bool) (*httpAccessLogDefinition, error) {
	tableNames, err := this.FindTableNames(db, day)
	if err != nil {
		return nil, err
	}

	var prefix = "edgeHTTPAccessLogs_" + day

	if len(tableNames) == 0 {
		if force {
			err := this.CreateTable(db, prefix)
			if err != nil {
				return nil, err
			}

			return &httpAccessLogDefinition{
				Name:          prefix,
				HasRemoteAddr: true,
				HasDomain:     true,
				Exists:        true,
			}, nil
		}

		return &httpAccessLogDefinition{
			Name:          prefix,
			HasRemoteAddr: true,
			HasDomain:     true,
			Exists:        false,
		}, nil
	}

	var lastTableName = tableNames[len(tableNames)-1]
	if !force || !accessLogEnableAutoPartial || accessLogRowsPerTable <= 0 {
		hasRemoteAddrField, hasDomainField, err := this.checkTableFields(db, lastTableName)
		if err != nil {
			return nil, err
		}
		return &httpAccessLogDefinition{
			Name:          lastTableName,
			HasRemoteAddr: hasRemoteAddrField,
			HasDomain:     hasDomainField,
			Exists:        true,
		}, nil
	}

	// 检查是否生成下个分表
	lastId, err := db.FindCol(0, "SELECT id FROM "+lastTableName+" ORDER BY id DESC LIMIT 1")
	if err != nil {
		return nil, err
	}

	if lastId != nil {
		var lastInt64Id = types.Int64(lastId)
		if accessLogRowsPerTable > 0 && lastInt64Id >= accessLogRowsPerTable {
			// create next partial table
			var nextTableName = ""
			if accessLogTableMainReg.MatchString(lastTableName) {
				nextTableName = prefix + "_0001"
			} else if accessLogTablePartialReg.MatchString(lastTableName) {
				var matches = accessLogTablePartialReg.FindStringSubmatch(lastTableName)
				if len(matches) < 3 {
					return nil, errors.New("fatal error: invalid 'accessLogTablePartialReg'")
				}
				var lastPartial = matches[2]
				nextTableName = prefix + "_" + fmt.Sprintf("%04d", types.Int(lastPartial)+1)
			} else {
				nextTableName = prefix + "_0001"
			}

			err = this.CreateTable(db, nextTableName)
			if err != nil {
				return nil, err
			}

			return &httpAccessLogDefinition{
				Name:          nextTableName,
				HasRemoteAddr: true,
				HasDomain:     true,
				Exists:        true,
			}, nil
		}
	}

	// 检查字段
	hasRemoteAddrField, hasDomainField, err := this.checkTableFields(db, lastTableName)
	if err != nil {
		return nil, err
	}
	return &httpAccessLogDefinition{
		Name:          lastTableName,
		HasRemoteAddr: hasRemoteAddrField,
		HasDomain:     hasDomainField,
		Exists:        true,
	}, nil
}

// TODO 考虑缓存检查结果
func (this *HTTPAccessLogManager) checkTableFields(db *dbs.DB, tableName string) (hasRemoteAddrField bool, hasDomainField bool, err error) {
	fields, _, err := db.FindPreparedOnes("SHOW FIELDS FROM " + tableName)
	if err != nil {
		return false, false, err
	}
	for _, field := range fields {
		var fieldName = field.GetString("Field")
		if strings.ToLower(fieldName) == strings.ToLower("remoteAddr") {
			hasRemoteAddrField = true
		}
		if strings.ToLower(fieldName) == "domain" {
			hasDomainField = true
		}
	}
	return
}

// 组合表格的缓存Key
func (this *HTTPAccessLogManager) composeTableCacheKey(dsn string, day string) string {
	// 注意：格式一定要固定，下面清除缓存的时候需要用到
	return dsn + "_" + day
}
