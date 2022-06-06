package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"strconv"
	"strings"
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

// EnableUserNode 启用条目
func (this *UserNodeDAO) EnableUserNode(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserNodeStateEnabled).
		Update()
	return err
}

// DisableUserNode 禁用条目
func (this *UserNodeDAO) DisableUserNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", UserNodeStateDisabled).
		Update()
	return err
}

// FindEnabledUserNode 查找启用中的条目
func (this *UserNodeDAO) FindEnabledUserNode(tx *dbs.Tx, id int64) (*UserNode, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", UserNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserNode), err
}

// FindUserNodeName 根据主键查找名称
func (this *UserNodeDAO) FindUserNodeName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindAllEnabledUserNodes 列出所有可用用户节点
func (this *UserNodeDAO) FindAllEnabledUserNodes(tx *dbs.Tx) (result []*UserNode, err error) {
	_, err = this.Query(tx).
		State(UserNodeStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledUserNodes 计算用户节点数量
func (this *UserNodeDAO) CountAllEnabledUserNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(UserNodeStateEnabled).
		Count()
}

// CountAllEnabledAndOnUserNodes 计算启用的用户节点数量
func (this *UserNodeDAO) CountAllEnabledAndOnUserNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(UserNodeStateEnabled).
		Attr("isOn", true).
		Count()
}

// ListEnabledUserNodes 列出单页的用户节点
func (this *UserNodeDAO) ListEnabledUserNodes(tx *dbs.Tx, offset int64, size int64) (result []*UserNode, err error) {
	_, err = this.Query(tx).
		State(UserNodeStateEnabled).
		Offset(offset).
		Limit(size).
		Desc("order").
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// FindEnabledUserNodeIdWithAddr 根据主机名和端口获取ID
func (this *UserNodeDAO) FindEnabledUserNodeIdWithAddr(tx *dbs.Tx, protocol string, host string, port int) (int64, error) {
	addr := maps.Map{
		"protocol":  protocol,
		"host":      host,
		"portRange": strconv.Itoa(port),
	}
	addrJSON, err := json.Marshal(addr)
	if err != nil {
		return 0, err
	}

	one, err := this.Query(tx).
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

// CreateUserNode 创建用户节点
func (this *UserNodeDAO) CreateUserNode(tx *dbs.Tx, name string, description string, httpJSON []byte, httpsJSON []byte, accessAddrsJSON []byte, isOn bool) (nodeId int64, err error) {
	uniqueId, err := this.GenUniqueId(tx)
	if err != nil {
		return 0, err
	}
	secret := rands.String(32)
	err = NewApiTokenDAO().CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleUser)
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
	err = this.Save(tx, op)
	if err != nil {
		return
	}

	return types.Int64(op.Id), nil
}

// UpdateUserNode 修改用户节点
func (this *UserNodeDAO) UpdateUserNode(tx *dbs.Tx, nodeId int64, name string, description string, httpJSON []byte, httpsJSON []byte, accessAddrsJSON []byte, isOn bool) error {
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

	err := this.Save(tx, op)
	return err
}

// FindEnabledUserNodeWithUniqueId 根据唯一ID获取节点信息
func (this *UserNodeDAO) FindEnabledUserNodeWithUniqueId(tx *dbs.Tx, uniqueId string) (*UserNode, error) {
	result, err := this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", UserNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*UserNode), err
}

// FindEnabledUserNodeIdWithUniqueId 根据唯一ID获取节点ID
func (this *UserNodeDAO) FindEnabledUserNodeIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", UserNodeStateEnabled).
		ResultPk().
		FindInt64Col(0)
}

// GenUniqueId 生成唯一ID
func (this *UserNodeDAO) GenUniqueId(tx *dbs.Tx) (string, error) {
	for {
		uniqueId := rands.HexString(32)
		ok, err := this.Query(tx).
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

// UpdateNodeStatus 更改节点状态
func (this *UserNodeDAO) UpdateNodeStatus(tx *dbs.Tx, nodeId int64, nodeStatus *nodeconfigs.NodeStatus) error {
	if nodeStatus == nil {
		return nil
	}

	nodeStatusJSON, err := json.Marshal(nodeStatus)
	if err != nil {
		return err
	}

	_, err = this.Query(tx).
		Pk(nodeId).
		Set("status", nodeStatusJSON).
		Update()
	return err
}

// CountAllLowerVersionNodes 计算所有节点中低于某个版本的节点数量
func (this *UserNodeDAO) CountAllLowerVersionNodes(tx *dbs.Tx, version string) (int64, error) {
	return this.Query(tx).
		State(UserNodeStateEnabled).
		Where("status IS NOT NULL").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("version", utils.VersionToLong(version)).
		Count()
}

// CountAllEnabledAndOnOfflineNodes 计算离线节点数量
func (this *UserNodeDAO) CountAllEnabledAndOnOfflineNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(UserNodeStateEnabled).
		Attr("isOn", true).
		Where("(status IS NULL OR JSON_EXTRACT(status, '$.updatedAt')<UNIX_TIMESTAMP()-60)").
		Count()
}

// CountAllEnabledUserNodesWithSSLPolicyIds 计算使用SSL策略的所有用户节点数量
func (this *UserNodeDAO) CountAllEnabledUserNodesWithSSLPolicyIds(tx *dbs.Tx, sslPolicyIds []int64) (count int64, err error) {
	if len(sslPolicyIds) == 0 {
		return
	}
	policyStringIds := []string{}
	for _, policyId := range sslPolicyIds {
		policyStringIds = append(policyStringIds, strconv.FormatInt(policyId, 10))
	}
	return this.Query(tx).
		State(UserNodeStateEnabled).
		Where("(FIND_IN_SET(JSON_EXTRACT(https, '$.sslPolicyRef.sslPolicyId'), :policyIds)) ").
		Param("policyIds", strings.Join(policyStringIds, ",")).
		Count()
}
