package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type UserBillDAO dbs.DAO

func NewUserBillDAO() *UserBillDAO {
	return dbs.NewDAO(&UserBillDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserBills",
			Model:  new(UserBill),
			PkName: "id",
		},
	}).(*UserBillDAO)
}

var SharedUserBillDAO *UserBillDAO

func init() {
	dbs.OnReady(func() {
		SharedUserBillDAO = NewUserBillDAO()
	})
}
