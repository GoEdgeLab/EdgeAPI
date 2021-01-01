package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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

var SharedNodeGrantDAO *NodeGrantDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeGrantDAO = NewNodeGrantDAO()
	})
}

// 启用条目
func (this *NodeGrantDAO) EnableNodeGrant(tx *dbs.Tx, id uint32) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", NodeGrantStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeGrantDAO) DisableNodeGrant(tx *dbs.Tx, id int64) (err error) {
	_, err = this.Query(tx).
		Pk(id).
		Set("state", NodeGrantStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *NodeGrantDAO) FindEnabledNodeGrant(tx *dbs.Tx, id int64) (*NodeGrant, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeGrantStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeGrant), err
}

// 根据主键查找名称
func (this *NodeGrantDAO) FindNodeGrantName(tx *dbs.Tx, id uint32) (string, error) {
	name, err := this.Query(tx).
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}

// 创建认证信息
func (this *NodeGrantDAO) CreateGrant(tx *dbs.Tx, adminId int64, name string, method string, username string, password string, privateKey string, description string, nodeId int64) (grantId int64, err error) {
	op := NewNodeGrantOperator()
	op.AdminId = adminId
	op.Name = name
	op.Method = method

	switch method {
	case "user":
		op.Username = username
		op.Password = password
		op.Su = false // TODO 需要做到前端可以配置
	case "privateKey":
		op.PrivateKey = privateKey
	}
	op.Description = description
	op.NodeId = nodeId
	op.State = NodeGrantStateEnabled
	err = this.Save(tx, op)
	return types.Int64(op.Id), err
}

// 修改认证信息
func (this *NodeGrantDAO) UpdateGrant(tx *dbs.Tx, grantId int64, name string, method string, username string, password string, privateKey string, description string, nodeId int64) error {
	if grantId <= 0 {
		return errors.New("invalid grantId")
	}

	op := NewNodeGrantOperator()
	op.Id = grantId
	op.Name = name
	op.Method = method

	switch method {
	case "user":
		op.Username = username
		op.Password = password
		op.Su = false // TODO 需要做到前端可以配置
	case "privateKey":
		op.PrivateKey = privateKey
	}
	op.Description = description
	op.NodeId = nodeId
	err := this.Save(tx, op)
	return err
}

// 计算所有认证信息数量
func (this *NodeGrantDAO) CountAllEnabledGrants(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NodeGrantStateEnabled).
		Count()
}

// 列出单页的认证信息
func (this *NodeGrantDAO) ListEnabledGrants(tx *dbs.Tx, offset int64, size int64) (result []*NodeGrant, err error) {
	_, err = this.Query(tx).
		State(NodeGrantStateEnabled).
		Offset(offset).
		Size(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 列出所有的认证信息
func (this *NodeGrantDAO) FindAllEnabledGrants(tx *dbs.Tx) (result []*NodeGrant, err error) {
	_, err = this.Query(tx).
		State(NodeGrantStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
