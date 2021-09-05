package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type ConnectivityResultDAO dbs.DAO

func NewConnectivityResultDAO() *ConnectivityResultDAO {
	return dbs.NewDAO(&ConnectivityResultDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeConnectivityResults",
			Model:  new(ConnectivityResult),
			PkName: "id",
		},
	}).(*ConnectivityResultDAO)
}

var SharedConnectivityResultDAO *ConnectivityResultDAO

func init() {
	dbs.OnReady(func() {
		SharedConnectivityResultDAO = NewConnectivityResultDAO()
	})
}
