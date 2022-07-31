package accounts

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	OrderMethodStateEnabled  = 1 // 已启用
	OrderMethodStateDisabled = 0 // 已禁用
)

type OrderMethodDAO dbs.DAO

func NewOrderMethodDAO() *OrderMethodDAO {
	return dbs.NewDAO(&OrderMethodDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeOrderMethods",
			Model:  new(OrderMethod),
			PkName: "id",
		},
	}).(*OrderMethodDAO)
}

var SharedOrderMethodDAO *OrderMethodDAO

func init() {
	dbs.OnReady(func() {
		SharedOrderMethodDAO = NewOrderMethodDAO()
	})
}
