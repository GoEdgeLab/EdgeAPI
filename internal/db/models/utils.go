package models

import (
	"github.com/iwind/TeaGo/dbs"
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
