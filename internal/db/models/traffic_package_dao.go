package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	TrafficPackageStateEnabled  = 1 // 已启用
	TrafficPackageStateDisabled = 0 // 已禁用
)

type TrafficPackageDAO dbs.DAO

func NewTrafficPackageDAO() *TrafficPackageDAO {
	return dbs.NewDAO(&TrafficPackageDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeTrafficPackages",
			Model:  new(TrafficPackage),
			PkName: "id",
		},
	}).(*TrafficPackageDAO)
}

var SharedTrafficPackageDAO *TrafficPackageDAO

func init() {
	dbs.OnReady(func() {
		SharedTrafficPackageDAO = NewTrafficPackageDAO()
	})
}

// EnableTrafficPackage 启用条目
func (this *TrafficPackageDAO) EnableTrafficPackage(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", TrafficPackageStateEnabled).
		Update()
	return err
}

// DisableTrafficPackage 禁用条目
func (this *TrafficPackageDAO) DisableTrafficPackage(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", TrafficPackageStateDisabled).
		Update()
	return err
}

// FindEnabledTrafficPackage 查找启用中的条目
func (this *TrafficPackageDAO) FindEnabledTrafficPackage(tx *dbs.Tx, id int64) (*TrafficPackage, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(TrafficPackageStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*TrafficPackage), err
}
