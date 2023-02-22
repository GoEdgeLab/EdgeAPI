package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ADPackageInstanceStateEnabled  = 1 // 已启用
	ADPackageInstanceStateDisabled = 0 // 已禁用
)

type ADPackageInstanceDAO dbs.DAO

func NewADPackageInstanceDAO() *ADPackageInstanceDAO {
	return dbs.NewDAO(&ADPackageInstanceDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeADPackageInstances",
			Model:  new(ADPackageInstance),
			PkName: "id",
		},
	}).(*ADPackageInstanceDAO)
}

var SharedADPackageInstanceDAO *ADPackageInstanceDAO

func init() {
	dbs.OnReady(func() {
		SharedADPackageInstanceDAO = NewADPackageInstanceDAO()
	})
}

// EnableADPackageInstance 启用条目
func (this *ADPackageInstanceDAO) EnableADPackageInstance(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ADPackageInstanceStateEnabled).
		Update()
	return err
}

// FindEnabledADPackageInstance 查找启用中的条目
func (this *ADPackageInstanceDAO) FindEnabledADPackageInstance(tx *dbs.Tx, id int64) (*ADPackageInstance, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(ADPackageInstanceStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ADPackageInstance), err
}
