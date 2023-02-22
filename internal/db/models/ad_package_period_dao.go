package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ADPackagePeriodStateEnabled  = 1 // 已启用
	ADPackagePeriodStateDisabled = 0 // 已禁用
)

type ADPackagePeriodDAO dbs.DAO

func NewADPackagePeriodDAO() *ADPackagePeriodDAO {
	return dbs.NewDAO(&ADPackagePeriodDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeADPackagePeriods",
			Model:  new(ADPackagePeriod),
			PkName: "id",
		},
	}).(*ADPackagePeriodDAO)
}

var SharedADPackagePeriodDAO *ADPackagePeriodDAO

func init() {
	dbs.OnReady(func() {
		SharedADPackagePeriodDAO = NewADPackagePeriodDAO()
	})
}

// EnableADPackagePeriod 启用条目
func (this *ADPackagePeriodDAO) EnableADPackagePeriod(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ADPackagePeriodStateEnabled).
		Update()
	return err
}

// DisableADPackagePeriod 禁用条目
func (this *ADPackagePeriodDAO) DisableADPackagePeriod(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ADPackagePeriodStateDisabled).
		Update()
	return err
}

// FindEnabledADPackagePeriod 查找启用中的条目
func (this *ADPackagePeriodDAO) FindEnabledADPackagePeriod(tx *dbs.Tx, id int64) (*ADPackagePeriod, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(ADPackagePeriodStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ADPackagePeriod), err
}
