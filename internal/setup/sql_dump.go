package setup

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"regexp"
	"strings"
)

var recordsTables = []*SQLRecordsTable{
	{
		TableName:    "edgeRegionCities",
		UniqueFields: []string{"name", "provinceId"},
	},
	{
		TableName:    "edgeRegionCountries",
		UniqueFields: []string{"name"},
	},
	{
		TableName:    "edgeRegionProvinces",
		UniqueFields: []string{"name", "countryId"},
	},
	{
		TableName:    "edgeRegionProviders",
		UniqueFields: []string{"name"},
	},
}

type SQLDump struct {
}

func NewSQLDump() *SQLDump {
	return &SQLDump{}
}

// 导出数据
func (this *SQLDump) Dump(db *dbs.DB) (result *SQLDumpResult, err error) {
	result = &SQLDumpResult{}

	tableNames, err := db.TableNames()
	if err != nil {
		return result, err
	}
	for _, tableName := range tableNames {
		// 忽略一些分表
		if strings.HasPrefix(tableName, "edgeHTTPAccessLogs_") {
			continue
		}

		table, err := db.FindFullTable(tableName)
		if err != nil {
			return nil, err
		}
		sqlTable := &SQLTable{
			Name:       table.Name,
			Engine:     table.Engine,
			Charset:    table.Collation,
			Definition: regexp.MustCompile(" AUTO_INCREMENT=\\d+").ReplaceAllString(table.Code, ""),
		}

		// 字段
		fields := []*SQLField{}
		for _, field := range table.Fields {
			fields = append(fields, &SQLField{
				Name:       field.Name,
				Definition: field.Definition(),
			})
		}
		sqlTable.Fields = fields

		// 索引
		indexes := []*SQLIndex{}
		for _, index := range table.Indexes {
			indexes = append(indexes, &SQLIndex{
				Name:       index.Name,
				Definition: index.Definition(),
			})
		}
		sqlTable.Indexes = indexes

		// Records
		records := []*SQLRecord{}
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
				}
				for k, v := range one {
					record.Values[k] = types.String(v)
				}
				records = append(records, record)
			}
		}
		sqlTable.Records = records

		result.Tables = append(result.Tables, sqlTable)
	}

	return
}

// 应用数据
func (this *SQLDump) Apply(db *dbs.DB, newResult *SQLDumpResult) (ops []string, err error) {
	currentResult, err := this.Dump(db)
	if err != nil {
		return nil, err
	}

	// 新增表格
	for _, newTable := range newResult.Tables {
		oldTable := currentResult.FindTable(newTable.Name)
		if oldTable == nil {
			ops = append(ops, "+ table "+newTable.Name)
			_, err = db.Exec(newTable.Definition)
			if err != nil {
				return nil, err
			}
		} else if oldTable.Definition != newTable.Definition {
			// 对比字段
			// +
			for _, newField := range newTable.Fields {
				oldField := oldTable.FindField(newField.Name)
				if oldField == nil {
					ops = append(ops, "+ "+newTable.Name+" "+newField.Name)
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " ADD `" + newField.Name + "` " + newField.Definition)
					if err != nil {
						return nil, err
					}
				} else if !newField.EqualDefinition(oldField.Definition) {
					ops = append(ops, "* "+newTable.Name+" "+newField.Name)
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " MODIFY `" + newField.Name + "` " + newField.Definition)
					if err != nil {
						return nil, err
					}
				}
			}

			// 对比索引
			// +
			for _, newIndex := range newTable.Indexes {
				oldIndex := oldTable.FindIndex(newIndex.Name)
				if oldIndex == nil {
					ops = append(ops, "+ index "+newTable.Name+" "+newIndex.Name)
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " ADD " + newIndex.Definition)
					if err != nil {
						return nil, err
					}
				} else if oldIndex.Definition != newIndex.Definition {
					ops = append(ops, "* index "+newTable.Name+" "+newIndex.Name)
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " DROP KEY " + newIndex.Name)
					if err != nil {
						return nil, err
					}
					_, err = db.Exec("ALTER TABLE " + newTable.Name + " ADD " + newIndex.Definition)
					if err != nil {
						return nil, err
					}
				}
			}

			// -
			for _, oldIndex := range oldTable.Indexes {
				newIndex := newTable.FindIndex(oldIndex.Name)
				if newIndex == nil {
					ops = append(ops, "- index "+oldTable.Name+" "+oldIndex.Name)
					_, err = db.Exec("ALTER TABLE " + oldTable.Name + " DROP KEY " + oldIndex.Name)
					if err != nil {
						return nil, err
					}
				}
			}

			// 对比字段
			// -
			for _, oldField := range oldTable.Fields {
				newField := newTable.FindField(oldField.Name)
				if newField == nil {
					ops = append(ops, "- field "+oldTable.Name+" "+oldField.Name)
					_, err = db.Exec("ALTER TABLE " + oldTable.Name + " DROP COLUMN `" + oldField.Name + "`")
					if err != nil {
						return nil, err
					}
				}
			}
		}

		// 对比记录
		// +
		for _, record := range newTable.Records {
			queryArgs := []string{}
			queryValues := []interface{}{}
			valueStrings := []string{}
			for _, field := range record.UniqueFields {
				queryArgs = append(queryArgs, field+"=?")
				queryValues = append(queryValues, record.Values[field])
				valueStrings = append(valueStrings, record.Values[field])
			}
			one, err := db.FindOne("SELECT * FROM "+newTable.Name+" WHERE "+strings.Join(queryArgs, " AND "), queryValues...)
			if err != nil {
				return nil, err
			}
			if one == nil {
				ops = append(ops, "+ record "+newTable.Name+" "+strings.Join(valueStrings, ", "))
				params := []string{}
				args := []string{}
				values := []interface{}{}
				for k, v := range record.Values {
					// ID需要保留，因为各个表格之间需要有对应关系
					params = append(params, "`"+k+"`")
					args = append(args, "?")
					values = append(values, v)
				}
				_, err = db.Exec("INSERT INTO "+newTable.Name+" ("+strings.Join(params, ", ")+") VALUES ("+strings.Join(args, ", ")+")", values...)
				if err != nil {
					return nil, err
				}
			} else if !record.ValuesEquals(one) {
				ops = append(ops, "* record "+newTable.Name+" "+strings.Join(valueStrings, ", "))
				args := []string{}
				values := []interface{}{}
				for k, v := range record.Values {
					if k == "id" {
						continue
					}
					args = append(args, k+"=?")
					values = append(values, v)
				}
				values = append(values, one.GetInt("id"))
				_, err = db.Exec("UPDATE "+newTable.Name+" SET "+strings.Join(args, ", ")+" WHERE id=?", values...)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// 减少表格
	// 由于我们不删除任何表格，所以这里什么都不做

	// 升级数据
	err = UpgradeSQLData(db)
	if err != nil {
		return nil, errors.New("upgrade data failed: " + err.Error())
	}

	return
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
