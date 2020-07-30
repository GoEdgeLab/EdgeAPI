package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
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
func (this *NodeDAO) DisableNode(id int64) (err error) {
	_, err = this.Query().
		Pk(id).
		Set("state", NodeStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *NodeDAO) FindEnabledNode(id int64) (*Node, error) {
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

// 创建节点
func (this *NodeDAO) CreateNode(name string, clusterId int64) (nodeId int64, err error) {
	op := NewNodeOperator()
	op.Name = name
	op.NodeId = rands.HexString(32)
	op.Secret = rands.String(32)
	op.ClusterId = clusterId
	op.IsOn = 1
	op.State = NodeStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}

	return types.Int64(op.Id), nil
}

// 修改节点
func (this *NodeDAO) UpdateNode(nodeId int64, name string, clusterId int64) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}
	op := NewNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.ClusterId = clusterId
	_, err := this.Save(op)
	return err
}

// 计算所有节点数量
func (this *NodeDAO) CountAllEnabledNodes() (int64, error) {
	return this.Query().
		State(NodeStateEnabled).
		Count()
}

// 列出单页节点
func (this *NodeDAO) ListEnabledNodes(offset int64, size int64) (result []*Node, err error) {
	_, err = this.Query().
		State(NodeStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
