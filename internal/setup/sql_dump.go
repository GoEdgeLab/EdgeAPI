package setup

import (
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/types"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

var recordsTables = []*SQLRecordsTable{
	{
		TableName:    "edgeRegionCountries",
		UniqueFields: []string{"name"},
		ExceptFields: []string{"customName", "customCodes"},
	},
	{
		TableName:    "edgeRegionProvinces",
		UniqueFields: []string{"name", "countryId"},
		ExceptFields: []string{"customName", "customCodes"},
	},
	{
		TableName:    "edgeRegionCities",
		UniqueFields: []string{"name", "provinceId"},
		ExceptFields: []string{"customName", "customCodes"},
	},
	{
		TableName:    "edgeRegionTowns",
		UniqueFields: []string{"name", "cityId"},
		ExceptFields: []string{"customName", "customCodes"},
	},
	{
		TableName:    "edgeRegionProviders",
		UniqueFields: []string{"name"},
		ExceptFields: []string{"customName", "customCodes"},
	},
	{
		TableName:    "edgeFormalClientSystems",
		UniqueFields: []string{"dataId"},
	},
	{
		TableName:    "edgeFormalClientBrowsers",
		UniqueFields: []string{"dataId"},
	},
}

type sqlItem struct {
	sqlString string
	args      []any
}

type SQLDump struct {
}

func NewSQLDump() *SQLDump {
	return &SQLDump{}
}

// Dump 导出数据
func (this *SQLDump) Dump(db *dbs.DB, includingRecords bool) (result *SQLDumpResult, err error) {
	result = &SQLDumpResult{}

	tableNames, err := db.TableNames()
	if err != nil {
		return result, err
	}

	fullTableMap, err := this.findFullTables(db, tableNames)
	if err != nil {
		return nil, err
	}

	var autoIncrementReg = regexp.MustCompile(` AUTO_INCREMENT=\d+`)

	for _, table := range fullTableMap {
		var tableName = table.Name

		// 忽略一些分表
		if strings.HasPrefix(strings.ToLower(tableName), strings.ToLower("edgeHTTPAccessLogs_")) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(tableName), strings.ToLower("edgeNSAccessLogs_")) {
			continue
		}

		var sqlTable = &SQLTable{
			Name:       table.Name,
			Engine:     table.Engine,
			Charset:    table.Collation,
			Definition: autoIncrementReg.ReplaceAllString(table.Code, ""),
		}

		// 字段
		var fields = []*SQLField{}
		for _, field := range table.Fields {
			fields = append(fields, &SQLField{
				Name:       field.Name,
				Definition: field.Definition(),
			})
		}
		sqlTable.Fields = fields

		// 索引
		var indexes = []*SQLIndex{}
		for _, index := range table.Indexes {
			indexes = append(indexes, &SQLIndex{
				Name:       index.Name,
				Definition: index.Definition(),
			})
		}
		sqlTable.Indexes = indexes

		// Records
		var records = []*SQLRecord{}
		if includingRecords {
			recordsTable := this.findRecordsTable(tableName)
			if recordsTable != nil {
				ones, _, err := db.FindOnes("SELECT * FROM " + tableName + " ORDER BY id ASC")
				if err != nil {
					return result, err
				}
				for _, one := range ones {
					record := &SQLRecord{
						Id:           one.GetInt64("id"),
						Values:       map[string]string{},
						UniqueFields: recordsTable.UniqueFields,
						ExceptFields: recordsTable.ExceptFields,
					}
					for k, v := range one {
						// 需要排除的字段
						if lists.ContainsString(record.ExceptFields, k) {
							continue
						}

						record.Values[k] = types.String(v)
					}
					records = append(records, record)
				}
			}
		}
		sqlTable.Records = records

		result.Tables = append(result.Tables, sqlTable)
	}

	return
}

// Apply 应用数据
func (this *SQLDump) Apply(db *dbs.DB, newResult *SQLDumpResult, showLog bool) (ops []string, err error) {
	// 设置Innodb事务提交模式
	{
		// 检查是否为root用户
		config, _ := db.Config()
		if config == nil {
			return nil, nil
		}
		dsnConfig, err := mysql.ParseDSN(config.Dsn)
		if err != nil || dsnConfig == nil {
			return nil, err
		}
		if dsnConfig.User == "root" {
			result, err := db.FindOne("SHOW VARIABLES WHERE variable_name='innodb_flush_log_at_trx_commit'")
			if err == nil && result != nil {
				var oldValue = result.GetInt("Value")
				if oldValue == 1 {
					_, _ = db.Exec("SET GLOBAL innodb_flush_log_at_trx_commit=2")
				}
			}
		}
	}

	// 执行队列
	var execQueue = make(chan *sqlItem, 256)

	var threads = 32
	var wg = sync.WaitGroup{}
	wg.Add(threads + 1 /** applyQueue **/)

	var applyOps []string
	var applyErr error
	go func() {
		defer wg.Done()
		defer close(execQueue)

		applyOps, applyErr = this.applyQueue(db, newResult, showLog, execQueue)
	}()

	var sqlErrors = []error{}
	var sqlErrLocker = &sync.Mutex{}
	for i := 0; i < threads; i++ {
		go func() {
			defer wg.Done()

			for item := range execQueue {
				_, err := db.Exec(item.sqlString, item.args...)
				if err != nil {
					sqlErrLocker.Lock()
					sqlErrors = append(sqlErrors, errors.New(item.sqlString+": "+err.Error()))
					sqlErrLocker.Unlock()
					break
				}
			}
		}()
	}
	wg.Wait()

	if applyErr != nil {
		return nil, applyErr
	}

	if len(sqlErrors) == 0 {
		// 升级数据
		err = UpgradeSQLData(db)
		if err != nil {
			return nil, errors.New("upgrade data failed: " + err.Error())
		}

		return applyOps, nil
	}

	return nil, sqlErrors[0]
}

func (this *SQLDump) applyQueue(db *dbs.DB, newResult *SQLDumpResult, showLog bool, queue chan *sqlItem) (ops []string, err error) {
	var execSQL = func(sqlString string, args ...any) {
		queue <- &sqlItem{
			sqlString: sqlString,
			args:      args,
		}
	}

	currentResult, err := this.Dump(db, false)
	if err != nil {
		return nil, err
	}

	// 新增表格
	for _, newTable := range newResult.Tables {
		var oldTable = currentResult.FindTable(newTable.Name)
		if oldTable == nil {
			var op = "+ table " + newTable.Name
			ops = append(ops, op)
			if showLog {
				fmt.Println(op)
			}
			if len(newTable.Records) == 0 {
				execSQL(newTable.Definition)
			} else {
				_, err = db.Exec(newTable.Definition)
				if err != nil {
					return nil, errors.New("'" + op + "' failed: " + err.Error())
				}
			}
		} else if oldTable.Definition != newTable.Definition {
			// 对比字段
			// +
			for _, newField := range newTable.Fields {
				var oldField = oldTable.FindField(newField.Name)
				if oldField == nil {
					var op = "+ " + newTable.Name + " " + newField.Name
					ops = append(ops, op)
					if showLog {
						fmt.Println(op)
					}
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " ADD `" + newField.Name + "` " + newField.Definition)
					if err != nil {
						return nil, errors.New("'" + op + "' failed: " + err.Error())
					}
				} else if !newField.EqualDefinition(oldField.Definition) {
					var op = "* " + newTable.Name + " " + newField.Name
					ops = append(ops, op)
					if showLog {
						fmt.Println(op)
					}
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " MODIFY `" + newField.Name + "` " + newField.Definition)
					if err != nil {
						return nil, errors.New("'" + op + "' failed: " + err.Error())
					}
				}
			}

			// 对比索引
			// +
			for _, newIndex := range newTable.Indexes {
				var oldIndex = oldTable.FindIndex(newIndex.Name)
				if oldIndex == nil {
					var op = "+ index " + newTable.Name + " " + newIndex.Name
					ops = append(ops, op)
					if showLog {
						fmt.Println(op)
					}
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " ADD " + newIndex.Definition)
					if err != nil {
						err = this.tryCreateIndex(err, db, newTable.Name, newIndex.Definition)
						if err != nil {
							return nil, errors.New("'" + op + "' failed: " + err.Error())
						}
					}
				} else if oldIndex.Definition != newIndex.Definition {
					var op = "* index " + newTable.Name + " " + newIndex.Name
					ops = append(ops, op)
					if showLog {
						fmt.Println(op)
					}
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " DROP KEY " + newIndex.Name)
					if err != nil {
						return nil, errors.New("'" + op + "' drop old key failed: " + err.Error())
					}
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " ADD " + newIndex.Definition)
					if err != nil {
						err = this.tryCreateIndex(err, db, newTable.Name, newIndex.Definition)
						if err != nil {
							return nil, errors.New("'" + op + "' failed: " + err.Error())
						}
					}
				}
			}

			// -
			for _, oldIndex := range oldTable.Indexes {
				var newIndex = newTable.FindIndex(oldIndex.Name)
				if newIndex == nil {
					var op = "- index " + oldTable.Name + " " + oldIndex.Name
					ops = append(ops, op)
					if showLog {
						fmt.Println(op)
					}
					_, err = db.Exec("ALTER TABLE " + oldTable.Name + " DROP KEY " + oldIndex.Name)
					if err != nil {
						return nil, errors.New("'" + op + "' failed: " + err.Error())
					}
				}
			}

			// 对比字段
			// -
			for _, oldField := range oldTable.Fields {
				var newField = newTable.FindField(oldField.Name)
				if newField == nil {
					var op = "- field " + oldTable.Name + " " + oldField.Name
					ops = append(ops, op)
					if showLog {
						fmt.Println(op)
					}
					_, err = db.Exec("ALTER TABLE " + oldTable.Name + " DROP COLUMN `" + oldField.Name + "`")
					if err != nil {
						return nil, errors.New("'" + op + "' failed: " + err.Error())
					}
				}
			}
		}

		// 对比记录
		// +
		for _, record := range newTable.Records {
			var queryArgs = []string{}
			var queryValues = []any{}
			var valueStrings = []string{}
			for _, field := range record.UniqueFields {
				queryArgs = append(queryArgs, field+"=?")
				queryValues = append(queryValues, record.Values[field])
				valueStrings = append(valueStrings, record.Values[field])
			}

			var recordId int64
			for field, recordValue := range record.Values {
				if field == "id" {
					recordId = types.Int64(recordValue)
					break
				}
			}

			queryValues = append(queryValues, recordId)
			one, err := db.FindOne("SELECT * FROM "+newTable.Name+" WHERE (("+strings.Join(queryArgs, " AND ")+") OR id=?)", queryValues...)
			if err != nil {
				return nil, err
			}
			if one == nil {
				ops = append(ops, "+ record "+newTable.Name+" "+strings.Join(valueStrings, ", "))
				if showLog {
					fmt.Println("+ record " + newTable.Name + " " + strings.Join(valueStrings, ", "))
				}
				var params = []string{}
				var args = []string{}
				var values = []any{}
				for k, v := range record.Values {
					// 需要排除的字段
					if lists.ContainsString(record.ExceptFields, k) {
						continue
					}

					// ID需要保留，因为各个表格之间需要有对应关系
					params = append(params, "`"+k+"`")
					args = append(args, "?")
					values = append(values, v)
				}

				execSQL("INSERT INTO "+newTable.Name+" ("+strings.Join(params, ", ")+") VALUES ("+strings.Join(args, ", ")+")", values...)
			} else if !record.ValuesEquals(one) {
				ops = append(ops, "* record "+newTable.Name+" "+strings.Join(valueStrings, ", "))
				if showLog {
					fmt.Println("* record " + newTable.Name + " " + strings.Join(valueStrings, ", "))
				}
				var args = []string{}
				var values = []any{}
				for k, v := range record.Values {
					if k == "id" {
						continue
					}

					// 需要排除的字段
					if lists.ContainsString(record.ExceptFields, k) {
						continue
					}

					args = append(args, k+"=?")
					values = append(values, v)
				}
				values = append(values, one.GetInt("id"))

				execSQL("UPDATE "+newTable.Name+" SET "+strings.Join(args, ", ")+" WHERE id=?", values...)
			}
		}
	}

	// 减少表格
	// 由于我们不删除任何表格，所以这里什么都不做

	return
}

// 查找所有表的完整信息
func (this *SQLDump) findFullTables(db *dbs.DB, tableNames []string) ([]*dbs.Table, error) {
	var fullTables = []*dbs.Table{}
	if len(tableNames) == 0 {
		return fullTables, nil
	}

	var locker = &sync.Mutex{}
	var queue = make(chan string, len(tableNames))
	for _, tableName := range tableNames {
		queue <- tableName
	}

	var wg = &sync.WaitGroup{}
	var concurrent = 8

	if runtime.NumCPU() > 4 {
		concurrent = 32
	}

	wg.Add(concurrent)
	var lastErr error
	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case tableName := <-queue:
					table, err := db.FindFullTable(tableName)
					if err != nil {
						locker.Lock()
						lastErr = err
						locker.Unlock()
						return
					}
					locker.Lock()
					table.Name = tableName
					fullTables = append(fullTables, table)
					locker.Unlock()
				default:
					return
				}
			}
		}()
	}
	wg.Wait()
	if lastErr != nil {
		return nil, lastErr
	}

	// 排序
	sort.Slice(fullTables, func(i, j int) bool {
		return fullTables[i].Name < fullTables[j].Name
	})

	return fullTables, nil
}

// 查找有记录的表
func (this *SQLDump) findRecordsTable(tableName string) *SQLRecordsTable {
	for _, table := range recordsTables {
		if table.TableName == tableName {
			return table
		}
	}
	return nil
}

// 创建索引
func (this *SQLDump) tryCreateIndex(err error, db *dbs.DB, tableName string, indexDefinition string) error {
	if err == nil {
		return nil
	}

	// 处理Duplicate entry
	if strings.Contains(err.Error(), "Error 1062: Duplicate entry") && (strings.HasSuffix(tableName, "Stats") || strings.HasSuffix(tableName, "Values")) {
		var tries = 5 // 尝试次数
		for i := 0; i < tries; i++ {
			_, err = db.Exec("TRUNCATE TABLE " + tableName)
			if err != nil {
				if i == tries-1 {
					return err
				}
				continue
			}
			_, err = db.Exec("ALTER TABLE " + tableName + " ADD " + indexDefinition)
			if err != nil {
				if i == tries-1 {
					return err
				}
			} else {
				return nil
			}
		}
	}

	return err
}
