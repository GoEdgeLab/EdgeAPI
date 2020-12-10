package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type ServerStatDAO dbs.DAO

func NewServerStatDAO() *ServerStatDAO {
	return dbs.NewDAO(&ServerStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerStats",
			Model:  new(ServerStat),
			PkName: "id",
		},
	}).(*ServerStatDAO)
}

var SharedServerStatDAO *ServerStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerStatDAO = NewServerStatDAO()
	})
}
