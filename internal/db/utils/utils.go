package dbutils

import (
	"github.com/iwind/TeaGo/dbs"
	"strings"
	"sync"
)

var SharedCacheLocker = sync.RWMutex{}

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

// QuoteLikeKeyword 处理关键词中的特殊字符
func QuoteLikeKeyword(keyword string) string {
	keyword = strings.ReplaceAll(keyword, "%", "\\%")
	keyword = strings.ReplaceAll(keyword, "_", "\\_")
	return keyword
}

func QuoteLike(keyword string) string {
	return "%" + QuoteLikeKeyword(keyword) + "%"
}

func QuoteLikePrefix(keyword string) string {
	return QuoteLikeKeyword(keyword) + "%"
}

func QuoteLikeSuffix(keyword string) string {
	return "%" + QuoteLikeKeyword(keyword)
}
