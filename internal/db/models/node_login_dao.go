package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
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

// 创建认证
func (this *NodeLoginDAO) CreateNodeLogin(nodeId int64, name string, loginType string, paramsJSON []byte) (loginId int64, err error) {
	login := NewNodeLoginOperator()
	login.NodeId = nodeId
	login.Name = name
	login.Type = loginType
	login.Params = string(paramsJSON)
	login.State = NodeLoginStateEnabled
	_, err = this.Save(login)
	return types.Int64(login.Id), err
}

// 修改认证
func (this *NodeLoginDAO) UpdateNodeLogin(loginId int64, name string, loginType string, paramsJSON []byte) error {
	if loginId <= 0 {
		return errors.New("invalid loginId")
	}
	login := NewNodeLoginOperator()
	login.Id = loginId
	login.Name = name
	login.Type = loginType
	login.Params = string(paramsJSON)
	_, err := this.Save(login)
	return err
}

// 查找认证
func (this *NodeLoginDAO) FindEnabledNodeLoginWithNodeId(nodeId int64) (*NodeLogin, error) {
	one, err := this.Query().
		Attr("nodeId", nodeId).
		State(NodeLoginStateEnabled).
		Find()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return one.(*NodeLogin), nil
}

// 禁用某个节点的认证
func (this *NodeLoginDAO) DisableNodeLogins(nodeId int64) error {
	_, err := this.Query().
		Attr("nodeId", nodeId).
		Set("state", NodeLoginStateDisabled).
		Update()
	return err
}
