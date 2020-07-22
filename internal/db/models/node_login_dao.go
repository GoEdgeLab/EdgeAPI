package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodeLoginStateEnabled  = 1 // 已启用
	NodeLoginStateDisabled = 0 // 已禁用
)

type NodeLoginDAO dbs.DAO

func NewNodeLoginDAO() *NodeLoginDAO {
	return dbs.NewDAO(&NodeLoginDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeLogins",
			Model:  new(NodeLogin),
			PkName: "id",
		},
	}).(*NodeLoginDAO)
}

var SharedNodeLoginDAO = NewNodeLoginDAO()

// 启用条目
func (this *NodeLoginDAO) EnableNodeLogin(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeLoginStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeLoginDAO) DisableNodeLogin(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeLoginStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *NodeLoginDAO) FindEnabledNodeLogin(id uint32) (*NodeLogin, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", NodeLoginStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeLogin), err
}

// 根据主键查找名称
func (this *NodeLoginDAO) FindNodeLoginName(id uint32) (string, error) {
	name, err := this.Query().
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}
