package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"strconv"
)

const (
	UserNodeStateEnabled  = 1 // 已启用
	UserNodeStateDisabled = 0 // 已禁用
)

type UserNodeDAO dbs.DAO

func NewUserNodeDAO() *UserNodeDAO {
	return dbs.NewDAO(&UserNodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeUserNodes",
			Model:  new(UserNode),
			PkName: "id",
		},
	}).(*UserNodeDAO)
}

var SharedUserNodeDAO *UserNodeDAO

func init() {
	dbs.OnReady(func() {
		SharedUserNodeDAO = NewUserNodeDAO()
	})
}

// 启用条目
func (this *UserNodeDAO) EnableUserNode(id uint32) error {
	_, err := this.Query().
		Pk(id).
		Set("state", UserNodeStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *UserNodeDAO) DisableUserNode(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", UserNodeStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *UserNodeDAO) FindEnabledUserNode(id int64) (*UserNode, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", UserNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserNode), err
}

// 根据主键查找名称
func (this *UserNodeDAO) FindUserNodeName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 列出所有可用用户节点
func (this *UserNodeDAO) FindAllEnabledUserNodes() (result []*UserNode, err error) {
	_, err = this.Query().
		State(UserNodeStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// 计算用户节点数量
func (this *UserNodeDAO) CountAllEnabledUserNodes() (int64, error) {
	return this.Query().
		State(UserNodeStateEnabled).
		Count()
}

// 列出单页的用户节点
func (this *UserNodeDAO) ListEnabledUserNodes(offset int64, size int64) (result []*UserNode, err error) {
	_, err = this.Query().
		State(UserNodeStateEnabled).
		Offset(offset).
		Limit(size).
		Desc("order").
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 根据主机名和端口获取ID
func (this *UserNodeDAO) FindEnabledUserNodeIdWithAddr(protocol string, host string, port int) (int64, error) {
	addr := maps.Map{
		"protocol":  protocol,
		"host":      host,
		"portRange": strconv.Itoa(port),
	}
	addrJSON, err := json.Marshal(addr)
	if err != nil {
		return 0, err
	}

	one, err := this.Query().
		State(UserNodeStateEnabled).
		Where("JSON_CONTAINS(accessAddrs, :addr)").
		Param("addr", string(addrJSON)).
		ResultPk().
		Find()
	if err != nil {
		return 0, err
	}
	if one == nil {
		return 0, nil
	}
	return int64(one.(*UserNode).Id), nil
}

// 创建用户节点
func (this *UserNodeDAO) CreateUserNode(name string, description string, httpJSON []byte, httpsJSON []byte, accessAddrsJSON []byte, isOn bool) (nodeId int64, err error) {
	uniqueId, err := this.genUniqueId()
	if err != nil {
		return 0, err
	}
	secret := rands.String(32)
	err = NewApiTokenDAO().CreateAPIToken(uniqueId, secret, NodeRoleUser)
	if err != nil {
		return
	}

	op := NewUserNodeOperator()
	op.IsOn = isOn
	op.UniqueId = uniqueId
	op.Secret = secret
	op.Name = name
	op.Description = description

	if len(httpJSON) > 0 {
		op.Http = httpJSON
	}
	if len(httpsJSON) > 0 {
		op.Https = httpsJSON
	}
	if len(accessAddrsJSON) > 0 {
		op.AccessAddrs = accessAddrsJSON
	}

	op.State = NodeStateEnabled
	err = this.Save(op)
	if err != nil {
		return
	}

	return types.Int64(op.Id), nil
}

// 修改用户节点
func (this *UserNodeDAO) UpdateUserNode(nodeId int64, name string, description string, httpJSON []byte, httpsJSON []byte, accessAddrsJSON []byte, isOn bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	op := NewUserNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.Description = description
	op.IsOn = isOn

	if len(httpJSON) > 0 {
		op.Http = httpJSON
	} else {
		op.Http = "null"
	}
	if len(httpsJSON) > 0 {
		op.Https = httpsJSON
	} else {
		op.Https = "null"
	}
	if len(accessAddrsJSON) > 0 {
		op.AccessAddrs = accessAddrsJSON
	} else {
		op.AccessAddrs = "null"
	}

	err := this.Save(op)
	return err
}

// 根据唯一ID获取节点信息
func (this *UserNodeDAO) FindEnabledUserNodeWithUniqueId(uniqueId string) (*UserNode, error) {
	result, err := this.Query().
		Attr("uniqueId", uniqueId).
		Attr("state", UserNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserNode), err
}

// 生成唯一ID
func (this *UserNodeDAO) genUniqueId() (string, error) {
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
