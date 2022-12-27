package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type UserEmailNotificationDAO dbs.DAO

func NewUserEmailNotificationDAO() *UserEmailNotificationDAO {
	return dbs.NewDAO(&UserEmailNotificationDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserEmailNotifications",
			Model:  new(UserEmailNotification),
			PkName: "id",
		},
	}).(*UserEmailNotificationDAO)
}

var SharedUserEmailNotificationDAO *UserEmailNotificationDAO

func init() {
	dbs.OnReady(func() {
		SharedUserEmailNotificationDAO = NewUserEmailNotificationDAO()
	})
}
