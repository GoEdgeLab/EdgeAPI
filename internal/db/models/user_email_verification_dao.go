package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type UserEmailVerificationDAO dbs.DAO

func NewUserEmailVerificationDAO() *UserEmailVerificationDAO {
	return dbs.NewDAO(&UserEmailVerificationDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserEmailVerifications",
			Model:  new(UserEmailVerification),
			PkName: "id",
		},
	}).(*UserEmailVerificationDAO)
}

var SharedUserEmailVerificationDAO *UserEmailVerificationDAO

func init() {
	dbs.OnReady(func() {
		SharedUserEmailVerificationDAO = NewUserEmailVerificationDAO()
	})
}
