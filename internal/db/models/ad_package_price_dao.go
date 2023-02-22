package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type ADPackagePriceDAO dbs.DAO

func NewADPackagePriceDAO() *ADPackagePriceDAO {
	return dbs.NewDAO(&ADPackagePriceDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeADPackagePrices",
			Model:  new(ADPackagePrice),
			PkName: "id",
		},
	}).(*ADPackagePriceDAO)
}

var SharedADPackagePriceDAO *ADPackagePriceDAO

func init() {
	dbs.OnReady(func() {
		SharedADPackagePriceDAO = NewADPackagePriceDAO()
	})
}
