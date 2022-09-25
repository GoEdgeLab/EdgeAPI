package dbutils

import (
	"github.com/iwind/TeaGo/dbs"
	"strings"
)

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

// SetGlobalVarMin 设置变量最小值
func SetGlobalVarMin(db *dbs.DB, variableName string, minValue int) error {
	result, err := db.FindOne("SHOW VARIABLES WHERE variable_name=?", variableName)
	if err != nil {
		return err
	}
	if len(result) == 0 {
		return nil
	}
	var oldValue = result.GetInt("Value")
	if oldValue > 0 /** 小于等于0通常表示不限制 **/ && oldValue < minValue {
		_, err = db.Exec("SET GLOBAL "+variableName+"=?", minValue)
		return err
	}
	return nil
}

// SetGlobalVarMax 设置变量最大值
func SetGlobalVarMax(db *dbs.DB, variableName string, maxValue int) error {
	result, err := db.FindOne("SHOW VARIABLES WHERE variable_name=?", variableName)
	if err != nil {
		return err
	}
	if len(result) == 0 {
		return nil
	}
	var oldValue = result.GetInt("Value")
	if oldValue > maxValue {
		_, err = db.Exec("SET GLOBAL "+variableName+"=?", maxValue)
		return err
	}
	return nil
}
