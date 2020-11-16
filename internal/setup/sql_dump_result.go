package setup

type SQLDumpResult struct {
	Tables []*SQLTable `json:"tables"`
}

func (this *SQLDumpResult) FindTable(tableName string) *SQLTable {
	for _, table := range this.Tables {
		if table.Name == tableName {
			return table
		}
	}
	return nil
}
