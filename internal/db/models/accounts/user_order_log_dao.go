package accounts

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type UserOrderLogDAO dbs.DAO

func NewUserOrderLogDAO() *UserOrderLogDAO {
	return dbs.NewDAO(&UserOrderLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserOrderLogs",
			Model:  new(UserOrderLog),
			PkName: "id",
		},
	}).(*UserOrderLogDAO)
}

var SharedUserOrderLogDAO *UserOrderLogDAO

func init() {
	dbs.OnReady(func() {
		SharedUserOrderLogDAO = NewUserOrderLogDAO()
	})
}
