package nameservers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NSDomainStateEnabled  = 1 // 已启用
	NSDomainStateDisabled = 0 // 已禁用
)

type NSDomainDAO dbs.DAO

func NewNSDomainDAO() *NSDomainDAO {
	return dbs.NewDAO(&NSDomainDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNSDomains",
			Model:  new(NSDomain),
			PkName: "id",
		},
	}).(*NSDomainDAO)
}

var SharedNSDomainDAO *NSDomainDAO

func init() {
	dbs.OnReady(func() {
		SharedNSDomainDAO = NewNSDomainDAO()
	})
}
