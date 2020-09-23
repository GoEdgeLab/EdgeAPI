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
	APINodeStateEnabled  = 1 // 已启用
	APINodeStateDisabled = 0 // 已禁用
)

type APINodeDAO dbs.DAO

func NewAPINodeDAO() *APINodeDAO {
	return dbs.NewDAO(&APINodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeAPINodes",
			Model:  new(APINode),
			PkName: "id",
		},
	}).(*APINodeDAO)
}

var SharedAPINodeDAO = NewAPINodeDAO()

// 启用条目
func (this *APINodeDAO) EnableAPINode(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", APINodeStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *APINodeDAO) DisableAPINode(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", APINodeStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *APINodeDAO) FindEnabledAPINode(id int64) (*APINode, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", APINodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*APINode), err
}

// 根据主键查找名称
func (this *APINodeDAO) FindAPINodeName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建API节点
func (this *APINodeDAO) CreateAPINode(name string, description string, host string, port int) (nodeId int64, err error) {
	uniqueId, err := this.genUniqueId()
	if err != nil {
		return 0, err
	}
	secret := rands.String(32)
	err = SharedApiTokenDAO.CreateAPIToken(uniqueId, secret, NodeRoleAPI)
	if err != nil {
		return
	}

	op := NewAPINodeOperator()
	op.IsOn = true
	op.UniqueId = uniqueId
	op.Secret = secret
	op.Name = name
	op.Description = description
	op.Host = host
	op.Port = port
	op.State = NodeStateEnabled
	_, err = this.Save(op)
	if err != nil {
		return
	}

	return types.Int64(op.Id), nil
}

// 修改API节点
func (this *APINodeDAO) UpdateAPINode(nodeId int64, name string, description string, host string, port int) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	op := NewAPINodeOperator()
	op.Id = nodeId
	op.Name = name
	op.Description = description
	op.Host = host
	op.Port = port
	_, err := this.Save(op)
	return err
}

// 列出所有可用API节点
func (this *APINodeDAO) FindAllEnabledAPINodes() (result []*APINode, err error) {
	_, err = this.Query().
		Attr("clusterId", 0). // 非集群专用
		State(APINodeStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 计算API节点数量
func (this *APINodeDAO) CountAllEnabledAPINodes() (int64, error) {
	return this.Query().
		State(APINodeStateEnabled).
		Count()
}

// 列出单页的API节点
func (this *APINodeDAO) ListEnabledAPINodes(offset int64, size int64) (result []*APINode, err error) {
	_, err = this.Query().
		Attr("clusterId", 0). // 非集群专用
		State(APINodeStateEnabled).
		Offset(offset).
		Limit(size).
		Desc("order").
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 生成唯一ID
func (this *APINodeDAO) genUniqueId() (string, error) {
	for {
		uniqueId := rands.HexString(32)
		ok, err := this.Query().
			Attr("uniqueId", uniqueId).
			Exist()
		if err != nil {
			return "", err
		}
		if ok {
			continue
		}
		return uniqueId, nil
	}
}
