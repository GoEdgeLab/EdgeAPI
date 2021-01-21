package models

import (
	"github.com/iwind/TeaGo/dbs"
	"sync"
)

var SharedCacheLocker = sync.RWMutex{}

// 处理JSON字节Slice
func JSONBytes(data []byte) []byte {
	if len(data) == 0 {
		return []byte("null")
	}
	return data
}

// 判断JSON是否不为空
func IsNotNull(data string) bool {
	return len(data) > 0 && data != "null"
}

// 构造Query
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
