package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type UserVerifyCodeDAO dbs.DAO

func NewUserVerifyCodeDAO() *UserVerifyCodeDAO {
	return dbs.NewDAO(&UserVerifyCodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserVerifyCodes",
			Model:  new(UserVerifyCode),
			PkName: "id",
		},
	}).(*UserVerifyCodeDAO)
}

var SharedUserVerifyCodeDAO *UserVerifyCodeDAO

func init() {
	dbs.OnReady(func() {
		SharedUserVerifyCodeDAO = NewUserVerifyCodeDAO()
	})
}
