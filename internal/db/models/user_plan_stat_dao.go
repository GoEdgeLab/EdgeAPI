package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type UserPlanStatDAO dbs.DAO

func NewUserPlanStatDAO() *UserPlanStatDAO {
	return dbs.NewDAO(&UserPlanStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserPlanStats",
			Model:  new(UserPlanStat),
			PkName: "id",
		},
	}).(*UserPlanStatDAO)
}

var SharedUserPlanStatDAO *UserPlanStatDAO

func init() {
	dbs.OnReady(func() {
		SharedUserPlanStatDAO = NewUserPlanStatDAO()
	})
}
