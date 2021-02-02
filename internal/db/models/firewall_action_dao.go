package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	FirewallActionStateEnabled  = 1 // 已启用
	FirewallActionStateDisabled = 0 // 已禁用
)

type FirewallActionDAO dbs.DAO

func NewFirewallActionDAO() *FirewallActionDAO {
	return dbs.NewDAO(&FirewallActionDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeFirewallActions",
			Model:  new(FirewallAction),
			PkName: "id",
		},
	}).(*FirewallActionDAO)
}

var SharedFirewallActionDAO *FirewallActionDAO

func init() {
	dbs.OnReady(func() {
		SharedFirewallActionDAO = NewFirewallActionDAO()
	})
}

// 启用条目
func (this *FirewallActionDAO) EnableFirewallAction(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", FirewallActionStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *FirewallActionDAO) DisableFirewallAction(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", FirewallActionStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *FirewallActionDAO) FindEnabledFirewallAction(tx *dbs.Tx, id uint32) (*FirewallAction, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", FirewallActionStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*FirewallAction), err
}

// 根据主键查找名称
func (this *FirewallActionDAO) FindFirewallActionName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}
