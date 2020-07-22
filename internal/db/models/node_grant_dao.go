package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodeGrantStateEnabled  = 1 // 已启用
	NodeGrantStateDisabled = 0 // 已禁用
)

type NodeGrantDAO dbs.DAO

func NewNodeGrantDAO() *NodeGrantDAO {
	return dbs.NewDAO(&NodeGrantDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeGrants",
			Model:  new(NodeGrant),
			PkName: "id",
		},
	}).(*NodeGrantDAO)
}

var SharedNodeGrantDAO = NewNodeGrantDAO()

// 启用条目
func (this *NodeGrantDAO) EnableNodeGrant(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeGrantStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeGrantDAO) DisableNodeGrant(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeGrantStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *NodeGrantDAO) FindEnabledNodeGrant(id uint32) (*NodeGrant, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", NodeGrantStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeGrant), err
}

// 根据主键查找名称
func (this *NodeGrantDAO) FindNodeGrantName(id uint32) (string, error) {
	name, err := this.Query().
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}
