package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	UserTrafficPackageStateEnabled  = 1 // 已启用
	UserTrafficPackageStateDisabled = 0 // 已禁用
)

type UserTrafficPackageDAO dbs.DAO

func NewUserTrafficPackageDAO() *UserTrafficPackageDAO {
	return dbs.NewDAO(&UserTrafficPackageDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserTrafficPackages",
			Model:  new(UserTrafficPackage),
			PkName: "id",
		},
	}).(*UserTrafficPackageDAO)
}

var SharedUserTrafficPackageDAO *UserTrafficPackageDAO

func init() {
	dbs.OnReady(func() {
		SharedUserTrafficPackageDAO = NewUserTrafficPackageDAO()
	})
}

// EnableUserTrafficPackage 启用条目
func (this *UserTrafficPackageDAO) EnableUserTrafficPackage(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTrafficPackageStateEnabled).
		Update()
	return err
}

// DisableUserTrafficPackage 禁用条目
func (this *UserTrafficPackageDAO) DisableUserTrafficPackage(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserTrafficPackageStateDisabled).
		Update()
	return err
}

// FindEnabledUserTrafficPackage 查找启用中的条目
func (this *UserTrafficPackageDAO) FindEnabledUserTrafficPackage(tx *dbs.Tx, id int64) (*UserTrafficPackage, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(UserTrafficPackageStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserTrafficPackage), err
}
