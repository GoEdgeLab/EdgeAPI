package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodeGroupStateEnabled  = 1 // 已启用
	NodeGroupStateDisabled = 0 // 已禁用
)

type NodeGroupDAO dbs.DAO

func NewNodeGroupDAO() *NodeGroupDAO {
	return dbs.NewDAO(&NodeGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeGroups",
			Model:  new(NodeGroup),
			PkName: "id",
		},
	}).(*NodeGroupDAO)
}

var SharedNodeGroupDAO = NewNodeGroupDAO()

// 启用条目
func (this *NodeGroupDAO) EnableNodeGroup(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeGroupStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeGroupDAO) DisableNodeGroup(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeGroupStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *NodeGroupDAO) FindEnabledNodeGroup(id uint32) (*NodeGroup, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", NodeGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeGroup), err
}

// 根据主键查找名称
func (this *NodeGroupDAO) FindNodeGroupName(id uint32) (string, error) {
	name, err := this.Query().
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}
