package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	UserScriptStateEnabled  = 1 // 已启用
	UserScriptStateDisabled = 0 // 已禁用
)

type UserScriptDAO dbs.DAO

func NewUserScriptDAO() *UserScriptDAO {
	return dbs.NewDAO(&UserScriptDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserScripts",
			Model:  new(UserScript),
			PkName: "id",
		},
	}).(*UserScriptDAO)
}

var SharedUserScriptDAO *UserScriptDAO

func init() {
	dbs.OnReady(func() {
		SharedUserScriptDAO = NewUserScriptDAO()
	})
}
