package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodeStateEnabled  = 1 // 已启用
	NodeStateDisabled = 0 // 已禁用
)

type NodeDAO dbs.DAO

func NewNodeDAO() *NodeDAO {
	return dbs.NewDAO(&NodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodes",
			Model:  new(Node),
			PkName: "id",
		},
	}).(*NodeDAO)
}

var SharedNodeDAO = NewNodeDAO()

// 启用条目
func (this *NodeDAO) EnableNode(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeDAO) DisableNode(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *NodeDAO) FindEnabledNode(id uint32) (*Node, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", NodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*Node), err
}

// 根据主键查找名称
func (this *NodeDAO) FindNodeName(id uint32) (string, error) {
	name, err := this.Query().
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}
