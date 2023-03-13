package setup

type SQLRecordsTable struct {
	TableName    string
	UniqueFields []string
	ExceptFields []string
	IgnoreId     bool // 是否可以排除ID
}
