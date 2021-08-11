package models

import (
	"errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	NodeLoginStateEnabled  = 1 // 已启用
	NodeLoginStateDisabled = 0 // 已禁用

	NodeLoginTypeSSH = "ssh"
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

var SharedNodeLoginDAO *NodeLoginDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeLoginDAO = NewNodeLoginDAO()
	})
}

// EnableNodeLogin 启用条目
func (this *NodeLoginDAO) EnableNodeLogin(tx *dbs.Tx, id uint32) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", NodeLoginStateEnabled).
		Update()
}

// DisableNodeLogin 禁用条目
func (this *NodeLoginDAO) DisableNodeLogin(tx *dbs.Tx, loginId int64) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(loginId).
		Set("state", NodeLoginStateDisabled).
		Update()
}

// FindEnabledNodeLogin 查找启用中的条目
func (this *NodeLoginDAO) FindEnabledNodeLogin(tx *dbs.Tx, id int64) (*NodeLogin, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeLoginStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeLogin), err
}

// FindNodeLoginName 根据主键查找名称
func (this *NodeLoginDAO) FindNodeLoginName(tx *dbs.Tx, id uint32) (string, error) {
	name, err := this.Query(tx).
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}

// CreateNodeLogin 创建认证
func (this *NodeLoginDAO) CreateNodeLogin(tx *dbs.Tx, role nodeconfigs.NodeRole, nodeId int64, name string, loginType string, paramsJSON []byte) (loginId int64, err error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}

	login := NewNodeLoginOperator()
	login.Role = role
	login.NodeId = nodeId
	login.Name = name
	login.Type = loginType
	login.Params = string(paramsJSON)
	login.State = NodeLoginStateEnabled
	err = this.Save(tx, login)
	return types.Int64(login.Id), err
}

// UpdateNodeLogin 修改认证
func (this *NodeLoginDAO) UpdateNodeLogin(tx *dbs.Tx, loginId int64, name string, loginType string, paramsJSON []byte) error {
	if loginId <= 0 {
		return errors.New("invalid loginId")
	}
	login := NewNodeLoginOperator()
	login.Id = loginId
	login.Name = name
	login.Type = loginType
	login.Params = string(paramsJSON)
	err := this.Save(tx, login)
	return err
}

// FindEnabledNodeLoginWithNodeId 查找认证
func (this *NodeLoginDAO) FindEnabledNodeLoginWithNodeId(tx *dbs.Tx, role nodeconfigs.NodeRole, nodeId int64) (*NodeLogin, error) {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	one, err := this.Query(tx).
		Attr("role", role).
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

// DisableNodeLogins 禁用某个节点的认证
func (this *NodeLoginDAO) DisableNodeLogins(tx *dbs.Tx, role nodeconfigs.NodeRole, nodeId int64) error {
	if len(role) == 0 {
		role = nodeconfigs.NodeRoleNode
	}
	_, err := this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Set("state", NodeLoginStateDisabled).
		Update()
	return err
}
