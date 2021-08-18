package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type NodeIPAddressLogDAO dbs.DAO

func NewNodeIPAddressLogDAO() *NodeIPAddressLogDAO {
	return dbs.NewDAO(&NodeIPAddressLogDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeIPAddressLogs",
			Model:  new(NodeIPAddressLog),
			PkName: "id",
		},
	}).(*NodeIPAddressLogDAO)
}

var SharedNodeIPAddressLogDAO *NodeIPAddressLogDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeIPAddressLogDAO = NewNodeIPAddressLogDAO()
	})
}
