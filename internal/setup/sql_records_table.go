package setup

type SQLRecordsTable struct {
	TableName    string
	UniqueFields []string
	ExceptFields []string
}
