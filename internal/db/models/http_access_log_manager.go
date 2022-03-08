// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package models

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// 访问日志的两个表格形式
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
		ones, columnNames, err := db.FindOnes(`SHOW TABLES LIKE '` + prefix + `'`)
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
	sort.Strings(results)

	return results, nil
}

// FindTables 读取数据库中某日所有日志表
func (this *HTTPAccessLogManager) FindTables(db *dbs.DB, day string) ([]*httpAccessLogDefinition, error) {
	var results = []*httpAccessLogDefinition{}
	var tableNames = []string{}

	// 需要防止用户设置了表名自动小写
	for _, prefix := range []string{"edgeHTTPAccessLogs_" + day + "%", "edgehttpaccesslogs_" + day + "%"} {
		ones, columnNames, err := db.FindOnes(`SHOW TABLES LIKE '` + prefix + `'`)
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

// FindTable 根据日期获取表名
// 表名组成
//   - PREFIX_DAY
//   - PREFIX_DAY_0001
func (this *HTTPAccessLogManager) FindTable(db *dbs.DB, day string, force bool) (*httpAccessLogDefinition, error) {
	this.locker.Lock()
	defer this.locker.Unlock()

	config, err := db.Config()
	if err != nil {
		return nil, err
	}
	var cacheKey = config.Dsn
	def, ok := this.currentTableMapping[cacheKey]
	if ok {
		return def, nil
	}

	def, err = this.findTableWithoutCache(db, day, force)
	if err != nil {
		return nil, err
	}

	this.currentTableMapping[cacheKey] = def
	return def, nil
}

// CreateTable 创建访问日志表格
func (this *HTTPAccessLogManager) CreateTable(db *dbs.DB, tableName string) error {
	_, err := db.Exec("CREATE TABLE `" + tableName + "` (\n  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n  `serverId` int(11) unsigned DEFAULT '0' COMMENT '服务ID',\n  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n  `status` int(3) unsigned DEFAULT '0' COMMENT '状态码',\n  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n  `content` json DEFAULT NULL COMMENT '日志内容',\n  `requestId` varchar(128) DEFAULT NULL COMMENT '请求ID',\n  `firewallPolicyId` int(11) unsigned DEFAULT '0' COMMENT 'WAF策略ID',\n  `firewallRuleGroupId` int(11) unsigned DEFAULT '0' COMMENT 'WAF分组ID',\n  `firewallRuleSetId` int(11) unsigned DEFAULT '0' COMMENT 'WAF集ID',\n  `firewallRuleId` int(11) unsigned DEFAULT '0' COMMENT 'WAF规则ID',\n  `remoteAddr` varchar(64) DEFAULT NULL COMMENT 'IP地址',\n  `domain` varchar(128) DEFAULT NULL COMMENT '域名',\n  `requestBody` mediumblob COMMENT '请求内容',\n  `responseBody` mediumblob COMMENT '响应内容',\n  PRIMARY KEY (`id`),\n  KEY `serverId` (`serverId`),\n  KEY `nodeId` (`nodeId`),\n  KEY `serverId_status` (`serverId`,`status`),\n  KEY `requestId` (`requestId`),\n  KEY `firewallPolicyId` (`firewallPolicyId`),\n  KEY `firewallRuleGroupId` (`firewallRuleGroupId`),\n  KEY `firewallRuleSetId` (`firewallRuleSetId`),\n  KEY `firewallRuleId` (`firewallRuleId`),\n  KEY `remoteAddr` (`remoteAddr`),\n  KEY `domain` (`domain`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='访问日志';")
	if err != nil {
		// 快速判断错误方法
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1050 { // Error 1050: Table 'xxx' already exists
			return nil
		}

		// 防止二次包装过程中错误丢失的保底错误判断方法
		if strings.Contains(err.Error(), "Error 1050") {
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
	delete(this.currentTableMapping, config.Dsn)
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
	if !force || !accessLogEnableAutoPartial {
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
		if lastInt64Id >= accessLogPartialRows {
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
	fields, _, err := db.FindOnes("SHOW FIELDS FROM " + tableName)
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
