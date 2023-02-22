package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ADPackageStateEnabled  = 1 // 已启用
	ADPackageStateDisabled = 0 // 已禁用
)

type ADPackageDAO dbs.DAO

func NewADPackageDAO() *ADPackageDAO {
	return dbs.NewDAO(&ADPackageDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeADPackages",
			Model:  new(ADPackage),
			PkName: "id",
		},
	}).(*ADPackageDAO)
}

var SharedADPackageDAO *ADPackageDAO

func init() {
	dbs.OnReady(func() {
		SharedADPackageDAO = NewADPackageDAO()
	})
}

// EnableADPackage 启用条目
func (this *ADPackageDAO) EnableADPackage(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ADPackageStateEnabled).
		Update()
	return err
}

// DisableADPackage 禁用条目
func (this *ADPackageDAO) DisableADPackage(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ADPackageStateDisabled).
		Update()
	return err
}

// FindEnabledADPackage 查找启用中的条目
func (this *ADPackageDAO) FindEnabledADPackage(tx *dbs.Tx, id int64) (*ADPackage, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(ADPackageStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ADPackage), err
}

// FindADPackageName 根据主键查找名称
func (this *ADPackageDAO) FindADPackageName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}
