package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type TrafficPackagePriceDAO dbs.DAO

func NewTrafficPackagePriceDAO() *TrafficPackagePriceDAO {
	return dbs.NewDAO(&TrafficPackagePriceDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeTrafficPackagePrices",
			Model:  new(TrafficPackagePrice),
			PkName: "id",
		},
	}).(*TrafficPackagePriceDAO)
}

var SharedTrafficPackagePriceDAO *TrafficPackagePriceDAO

func init() {
	dbs.OnReady(func() {
		SharedTrafficPackagePriceDAO = NewTrafficPackagePriceDAO()
	})
}
