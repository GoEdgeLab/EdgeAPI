package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodeActionStateEnabled  = 1 // 已启用
	NodeActionStateDisabled = 0 // 已禁用
)

type NodeActionDAO dbs.DAO

func NewNodeActionDAO() *NodeActionDAO {
	return dbs.NewDAO(&NodeActionDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeActions",
			Model:  new(NodeAction),
			PkName: "id",
		},
	}).(*NodeActionDAO)
}

var SharedNodeActionDAO *NodeActionDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeActionDAO = NewNodeActionDAO()
	})
}

// EnableNodeAction 启用条目
func (this *NodeActionDAO) EnableNodeAction(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeActionStateEnabled).
		Update()
	return err
}

// DisableNodeAction 禁用条目
func (this *NodeActionDAO) DisableNodeAction(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeActionStateDisabled).
		Update()
	return err
}

// FindEnabledNodeAction 查找启用中的条目
func (this *NodeActionDAO) FindEnabledNodeAction(tx *dbs.Tx, id int64) (*NodeAction, error) {
	result, err := this.Query(tx).
		Pk(id).
		State(NodeActionStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeAction), err
}
