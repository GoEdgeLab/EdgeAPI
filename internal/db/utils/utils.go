package dbutils

import (
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"net"
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

// IsLocalAddr 是否为本地数据库
func IsLocalAddr(addr string) bool {
	var host = addr
	if strings.Contains(addr, ":") {
		host, _, _ = net.SplitHostPort(addr)
		if len(host) == 0 {
			host = addr
		}
	}

	if host == "127.0.0.1" || host == "::1" || host == "localhost" {
		return true
	}

	interfaceAddrs, _ := net.InterfaceAddrs()
	for _, interfaceAddr := range interfaceAddrs {
		if strings.HasPrefix(interfaceAddr.String(), host+"/") {
			return true
		}
	}
	return false
}

// MySQLVersion 读取当前MySQL版本
func MySQLVersion() (version string, err error) {
	db, err := dbs.Default()
	if err != nil {
		return "", err
	}
	result, err := db.FindCol(0, "SELECT VERSION()")
	if err != nil {
		return "", err
	}
	version = types.String(result)
	var suffixIndex = strings.Index(version, "-")
	if suffixIndex > 0 {
		version = version[:suffixIndex]
	}
	return
}

func MySQLVersionFrom8() (bool, error) {
	version, err := MySQLVersion()
	if err != nil {
		return false, err
	}
	if len(version) == 0 {
		return false, nil
	}
	var dotIndex = strings.Index(version, ".")
	if dotIndex > 0 {
		return types.Int(version[:dotIndex]) >= 8, nil
	}
	return false, nil
}
