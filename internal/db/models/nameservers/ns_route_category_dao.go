package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSRouteCategoryStateEnabled  = 1 // 已启用
	NSRouteCategoryStateDisabled = 0 // 已禁用
)

type NSRouteCategoryDAO dbs.DAO

func NewNSRouteCategoryDAO() *NSRouteCategoryDAO {
	return dbs.NewDAO(&NSRouteCategoryDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSRouteCategories",
			Model:  new(NSRouteCategory),
			PkName: "id",
		},
	}).(*NSRouteCategoryDAO)
}

var SharedNSRouteCategoryDAO *NSRouteCategoryDAO

func init() {
	dbs.OnReady(func() {
		SharedNSRouteCategoryDAO = NewNSRouteCategoryDAO()
	})
}
