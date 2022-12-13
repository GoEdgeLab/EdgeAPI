package models

import (
	"github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strings"
	"sync"
)

// SharedCacheLocker 缓存专用Locker
var SharedCacheLocker = sync.RWMutex{}

// JSONBytes 处理JSON字节Slice
func JSONBytes(data []byte) []byte {
	if len(data) == 0 {
		return []byte("null")
	}
	return data
}

// IsNotNull 判断JSON是否不为空
func IsNotNull(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	if len(data) == 4 && string(data) == "null" {
		return false
	}
	return true
}

// IsNull 判断JSON是否为空
func IsNull(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	if len(data) == 4 && string(data) == "null" {
		return true
	}
	return false
}

// NewQuery 构造Query
func NewQuery(tx *dbs.Tx, dao dbs.DAOWrapper, adminId int64, userId int64) *dbs.Query {
	query := dao.Object().Query(tx)
	if adminId > 0 {
		//query.Attr("adminId", adminId)
	}
	if userId > 0 {
		query.Attr("userId", userId)
	}
	return query
}

// CheckSQLErrCode 检查数据库错误代码
func CheckSQLErrCode(err error, code uint16) bool {
	if err == nil {
		return false
	}

	// 快速判断错误方法
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok && mysqlErr.Number == code { // Error 1050: Table 'xxx' already exists
		return true
	}

	// 防止二次包装过程中错误丢失的保底错误判断方法
	if strings.Contains(err.Error(), "Error "+types.String(code)) {
		return true
	}

	return false
}

// CheckSQLDuplicateErr 检查Duplicate错误
func CheckSQLDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	return CheckSQLErrCode(err, 1062)
}
