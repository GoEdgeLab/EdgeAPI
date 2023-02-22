package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	UserADInstanceStateEnabled  = 1 // 已启用
	UserADInstanceStateDisabled = 0 // 已禁用
)

type UserADInstanceDAO dbs.DAO

func NewUserADInstanceDAO() *UserADInstanceDAO {
	return dbs.NewDAO(&UserADInstanceDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserADInstances",
			Model:  new(UserADInstance),
			PkName: "id",
		},
	}).(*UserADInstanceDAO)
}

var SharedUserADInstanceDAO *UserADInstanceDAO

func init() {
	dbs.OnReady(func() {
		SharedUserADInstanceDAO = NewUserADInstanceDAO()
	})
}

// EnableUserADInstance 启用条目
func (this *UserADInstanceDAO) EnableUserADInstance(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserADInstanceStateEnabled).
		Update()
	return err
}
