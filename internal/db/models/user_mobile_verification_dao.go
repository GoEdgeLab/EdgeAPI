package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type UserMobileVerificationDAO dbs.DAO

func NewUserMobileVerificationDAO() *UserMobileVerificationDAO {
	return dbs.NewDAO(&UserMobileVerificationDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserMobileVerifications",
			Model:  new(UserMobileVerification),
			PkName: "id",
		},
	}).(*UserMobileVerificationDAO)
}

var SharedUserMobileVerificationDAO *UserMobileVerificationDAO

func init() {
	dbs.OnReady(func() {
		SharedUserMobileVerificationDAO = NewUserMobileVerificationDAO()
	})
}
