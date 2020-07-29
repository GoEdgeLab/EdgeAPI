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

var SharedNodeGrantDAO = NewNodeGrantDAO()

// 启用条目
func (this *NodeGrantDAO) EnableNodeGrant(id uint32) (rowsAffected int64, err error) {
	return this.Query().
		Pk(id).
		Set("state", NodeGrantStateEnabled).
		Update()
}

// 禁用条目
func (this *NodeGrantDAO) DisableNodeGrant(id int64) (err error) {
	_, err = this.Query().
		Pk(id).
		Set("state", NodeGrantStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *NodeGrantDAO) FindEnabledNodeGrant(id int64) (*NodeGrant, error) {
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

// 创建认证信息
func (this *NodeGrantDAO) CreateGrant(name string, method string, username string, password string, privateKey string, description string, nodeId int64) (grantId int64, err error) {
	op := NewNodeGrantOperator()
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
	_, err = this.Save(op)
	return types.Int64(op.Id), err
}

// 修改认证信息
func (this *NodeGrantDAO) UpdateGrant(grantId int64, name string, method string, username string, password string, privateKey string, description string, nodeId int64) error {
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
	_, err := this.Save(op)
	return err
}

// 计算所有认证信息数量
func (this *NodeGrantDAO) CountAllEnabledGrants() (int64, error) {
	return this.Query().
		State(NodeGrantStateEnabled).
		Count()
}

// 列出单页的认证信息
func (this *NodeGrantDAO) ListEnabledGrants(offset int64, size int64) (result []*NodeGrant, err error) {
	_, err = this.Query().
		State(NodeGrantStateEnabled).
		Offset(offset).
		Size(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 列出所有的认证信息
func (this *NodeGrantDAO) FindAllEnabledGrants() (result []*NodeGrant, err error) {
	_, err = this.Query().
		State(NodeGrantStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
