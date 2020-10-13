package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type VersionDAO dbs.DAO

func NewVersionDAO() *VersionDAO {
	return dbs.NewDAO(&VersionDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeVersions",
			Model:  new(Version),
			PkName: "id",
		},
	}).(*VersionDAO)
}

var SharedVersionDAO *VersionDAO

func init() {
	dbs.OnReady(func() {
		SharedVersionDAO = NewVersionDAO()
	})
}
