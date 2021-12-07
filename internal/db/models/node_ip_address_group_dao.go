package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

type NodeIPAddressGroupDAO dbs.DAO

func NewNodeIPAddressGroupDAO() *NodeIPAddressGroupDAO {
	return dbs.NewDAO(&NodeIPAddressGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeIPAddressGroups",
			Model:  new(NodeIPAddressGroup),
			PkName: "id",
		},
	}).(*NodeIPAddressGroupDAO)
}

var SharedNodeIPAddressGroupDAO *NodeIPAddressGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeIPAddressGroupDAO = NewNodeIPAddressGroupDAO()
	})
}

// FindNodeIPAddressGroupName 根据主键查找名称
func (this *NodeIPAddressGroupDAO) FindNodeIPAddressGroupName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateGroup 创建分组
func (this *NodeIPAddressGroupDAO) CreateGroup(tx *dbs.Tx, name string, value string) (int64, error) {
	var op = NewNodeIPAddressGroupOperator()
	op.Name = name
	op.Value = value
	return this.SaveInt64(tx, op)
}
