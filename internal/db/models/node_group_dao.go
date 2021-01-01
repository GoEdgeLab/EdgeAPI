package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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

var SharedNodeGroupDAO *NodeGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeGroupDAO = NewNodeGroupDAO()
	})
}

// 启用条目
func (this *NodeGroupDAO) EnableNodeGroup(tx *dbs.Tx, id int64) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", NodeGroupStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeGroupDAO) DisableNodeGroup(tx *dbs.Tx, id int64) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", NodeGroupStateDisabled).
		Update()
}

// 查找启用中的条目
func (this *NodeGroupDAO) FindEnabledNodeGroup(tx *dbs.Tx, id int64) (*NodeGroup, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeGroup), err
}

// 根据主键查找名称
func (this *NodeGroupDAO) FindNodeGroupName(tx *dbs.Tx, id int64) (string, error) {
	name, err := this.Query(tx).
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}

// 创建分组
func (this *NodeGroupDAO) CreateNodeGroup(tx *dbs.Tx, clusterId int64, name string) (int64, error) {
	op := NewNodeGroupOperator()
	op.ClusterId = clusterId
	op.Name = name
	op.State = NodeGroupStateEnabled
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改分组
func (this *NodeGroupDAO) UpdateNodeGroup(tx *dbs.Tx, groupId int64, name string) error {
	if groupId <= 0 {
		return errors.New("invalid groupId")
	}
	op := NewNodeGroupOperator()
	op.Id = groupId
	op.Name = name
	err := this.Save(tx, op)
	return err
}

// 查询所有分组
func (this *NodeGroupDAO) FindAllEnabledGroupsWithClusterId(tx *dbs.Tx, clusterId int64) (result []*NodeGroup, err error) {
	_, err = this.Query(tx).
		State(NodeGroupStateEnabled).
		Attr("clusterId", clusterId).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 保存排序
func (this *NodeGroupDAO) UpdateGroupOrders(tx *dbs.Tx, groupIds []int64) error {
	for index, groupId := range groupIds {
		_, err := this.Query(tx).
			Pk(groupId).
			Set("order", len(groupIds)-index).
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}
