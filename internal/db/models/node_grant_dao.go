package models

import (
	"errors"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
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

// EnableNodeGrant 启用条目
func (this *NodeGrantDAO) EnableNodeGrant(tx *dbs.Tx, id uint32) (rowsAffected int64, err error) {
	return this.Query(tx).
		Pk(id).
		Set("state", NodeGrantStateEnabled).
		Update()
}

// DisableNodeGrant 禁用条目
func (this *NodeGrantDAO) DisableNodeGrant(tx *dbs.Tx, id int64) (err error) {
	_, err = this.Query(tx).
		Pk(id).
		Set("state", NodeGrantStateDisabled).
		Update()
	return err
}

// FindEnabledNodeGrant 查找启用中的条目
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

// FindNodeGrantName 根据主键查找名称
func (this *NodeGrantDAO) FindNodeGrantName(tx *dbs.Tx, id uint32) (string, error) {
	name, err := this.Query(tx).
		Pk(id).
		Result("name").
		FindCol("")
	return name.(string), err
}

// CreateGrant 创建认证信息
func (this *NodeGrantDAO) CreateGrant(tx *dbs.Tx, adminId int64, name string, method string, username string, password string, privateKey string, passphrase string, description string, nodeId int64, su bool) (grantId int64, err error) {
	var op = NewNodeGrantOperator()
	op.AdminId = adminId
	op.Name = name
	op.Method = method

	switch method {
	case "user":
		op.Username = username
		op.Password = password
	case "privateKey":
		op.Username = username
		op.PrivateKey = privateKey
		op.Passphrase = passphrase
	}
	op.Su = su
	op.Description = description
	op.NodeId = nodeId
	op.State = NodeGrantStateEnabled
	err = this.Save(tx, op)
	return types.Int64(op.Id), err
}

// UpdateGrant 修改认证信息
func (this *NodeGrantDAO) UpdateGrant(tx *dbs.Tx, grantId int64, name string, method string, username string, password string, privateKey string, passphrase string, description string, nodeId int64, su bool) error {
	if grantId <= 0 {
		return errors.New("invalid grantId")
	}

	var op = NewNodeGrantOperator()
	op.Id = grantId
	op.Name = name
	op.Method = method

	switch method {
	case "user":
		op.Username = username
		op.Password = password
	case "privateKey":
		op.Username = username
		op.PrivateKey = privateKey
		op.Passphrase = passphrase
	}
	op.Su = su
	op.Description = description
	op.NodeId = nodeId
	err := this.Save(tx, op)
	return err
}

// CountAllEnabledGrants 计算所有认证信息数量
func (this *NodeGrantDAO) CountAllEnabledGrants(tx *dbs.Tx, keyword string) (int64, error) {
	query := this.Query(tx).
		State(NodeGrantStateEnabled)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR username LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	return query.Count()
}

// ListEnabledGrants 列出单页的认证信息
func (this *NodeGrantDAO) ListEnabledGrants(tx *dbs.Tx, keyword string, offset int64, size int64) (result []*NodeGrant, err error) {
	query := this.Query(tx).
		State(NodeGrantStateEnabled)
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword OR username LIKE :keyword OR description LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	_, err = query.
		Offset(offset).
		Size(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledGrants 列出所有的认证信息
func (this *NodeGrantDAO) FindAllEnabledGrants(tx *dbs.Tx) (result []*NodeGrant, err error) {
	_, err = this.Query(tx).
		State(NodeGrantStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}
