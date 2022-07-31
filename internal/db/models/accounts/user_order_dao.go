package accounts

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	UserOrderStateEnabled  = 1 // 已启用
	UserOrderStateDisabled = 0 // 已禁用
)

type UserOrderDAO dbs.DAO

func NewUserOrderDAO() *UserOrderDAO {
	return dbs.NewDAO(&UserOrderDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserOrders",
			Model:  new(UserOrder),
			PkName: "id",
		},
	}).(*UserOrderDAO)
}

var SharedUserOrderDAO *UserOrderDAO

func init() {
	dbs.OnReady(func() {
		SharedUserOrderDAO = NewUserOrderDAO()
	})
}
