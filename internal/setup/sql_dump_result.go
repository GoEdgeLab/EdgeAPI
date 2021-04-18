package setup

import "strings"

type SQLDumpResult struct {
	Tables []*SQLTable `json:"tables"`
}

func (this *SQLDumpResult) FindTable(tableName string) *SQLTable {
	for _, table := range this.Tables {
		if strings.ToLower(table.Name) == strings.ToLower(tableName) {
			return table
		}
	}
	return nil
}
