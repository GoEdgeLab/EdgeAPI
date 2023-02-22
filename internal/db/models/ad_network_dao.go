package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	ADNetworkStateEnabled  = 1 // 已启用
	ADNetworkStateDisabled = 0 // 已禁用
)

type ADNetworkDAO dbs.DAO

func NewADNetworkDAO() *ADNetworkDAO {
	return dbs.NewDAO(&ADNetworkDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeADNetworks",
			Model:  new(ADNetwork),
			PkName: "id",
		},
	}).(*ADNetworkDAO)
}

var SharedADNetworkDAO *ADNetworkDAO

func init() {
	dbs.OnReady(func() {
		SharedADNetworkDAO = NewADNetworkDAO()
	})
}

// EnableADNetwork 启用条目
func (this *ADNetworkDAO) EnableADNetwork(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ADNetworkStateEnabled).
		Update()
	return err
}

// DisableADNetwork 禁用条目
func (this *ADNetworkDAO) DisableADNetwork(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", ADNetworkStateDisabled).
		Update()
	return err
}

// FindEnabledADNetwork 查找启用中的条目
func (this *ADNetworkDAO) FindEnabledADNetwork(tx *dbs.Tx, id int64) (*ADNetwork, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(ADNetworkStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*ADNetwork), err
}

// FindADNetworkName 根据主键查找名称
func (this *ADNetworkDAO) FindADNetworkName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}
