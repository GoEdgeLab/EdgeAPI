package models

import (
	"encoding/json"
	"errors"
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

var SharedAPINodeDAO *APINodeDAO

func init() {
	dbs.OnReady(func() {
		SharedAPINodeDAO = NewAPINodeDAO()
	})
}

// EnableAPINode 启用条目
func (this *APINodeDAO) EnableAPINode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", APINodeStateEnabled).
		Update()
	return err
}

// DisableAPINode 禁用条目
func (this *APINodeDAO) DisableAPINode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", APINodeStateDisabled).
		Update()
	return err
}

// FindEnabledAPINode 查找启用中的条目
func (this *APINodeDAO) FindEnabledAPINode(tx *dbs.Tx, id int64) (*APINode, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", APINodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*APINode), err
}

// FindEnabledAPINodeWithUniqueIdAndSecret 根据ID和Secret查找节点
func (this *APINodeDAO) FindEnabledAPINodeWithUniqueIdAndSecret(tx *dbs.Tx, uniqueId string, secret string) (*APINode, error) {
	one, err := this.Query(tx).
		State(APINodeStateEnabled).
		Attr("uniqueId", uniqueId).
		Attr("secret", secret).
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*APINode), nil
}

// FindAPINodeName 根据主键查找名称
func (this *APINodeDAO) FindAPINodeName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// CreateAPINode 创建API节点
func (this *APINodeDAO) CreateAPINode(tx *dbs.Tx, name string, description string, httpJSON []byte, httpsJSON []byte, restIsOn bool, restHTTPJSON []byte, restHTTPSJSON []byte, accessAddrsJSON []byte, isOn bool) (nodeId int64, err error) {
	uniqueId, err := this.genUniqueId(tx)
	if err != nil {
		return 0, err
	}
	secret := rands.String(32)
	err = NewApiTokenDAO().CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleAPI)
	if err != nil {
		return
	}

	op := NewAPINodeOperator()
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
	op.RestIsOn = restIsOn
	if len(restHTTPJSON) > 0 {
		op.RestHTTP = restHTTPJSON
	}
	if len(restHTTPSJSON) > 0 {
		op.RestHTTPS = restHTTPSJSON
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

// UpdateAPINode 修改API节点
func (this *APINodeDAO) UpdateAPINode(tx *dbs.Tx, nodeId int64, name string, description string, httpJSON []byte, httpsJSON []byte, restIsOn bool, restHTTPJSON []byte, restHTTPSJSON []byte, accessAddrsJSON []byte, isOn bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	op := NewAPINodeOperator()
	op.Id = nodeId
	op.Name = name
	op.Description = description
	op.IsOn = isOn

	if len(httpJSON) > 0 {
		op.Http = httpJSON
	} else {
		op.Http = "{}"
	}
	if len(httpsJSON) > 0 {
		op.Https = httpsJSON
	} else {
		op.Https = "{}"
	}
	op.RestIsOn = restIsOn
	if len(restHTTPJSON) > 0 {
		op.RestHTTP = restHTTPJSON
	} else {
		op.RestHTTP = "{}"
	}
	if len(restHTTPSJSON) > 0 {
		op.RestHTTPS = restHTTPSJSON
	} else {
		op.RestHTTPS = "{}"
	}
	if len(accessAddrsJSON) > 0 {
		op.AccessAddrs = accessAddrsJSON
	} else {
		op.AccessAddrs = "[]"
	}

	err := this.Save(tx, op)
	return err
}

// FindAllEnabledAPINodes 列出所有可用API节点
func (this *APINodeDAO) FindAllEnabledAPINodes(tx *dbs.Tx) (result []*APINode, err error) {
	_, err = this.Query(tx).
		Attr("clusterId", 0). // 非集群专用
		State(APINodeStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledAndOnAPINodes 列出所有可用而且启用的API节点
func (this *APINodeDAO) FindAllEnabledAndOnAPINodes(tx *dbs.Tx) (result []*APINode, err error) {
	_, err = this.Query(tx).
		Attr("clusterId", 0). // 非集群专用
		Attr("isOn", true).
		State(APINodeStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledAPINodes 计算API节点数量
func (this *APINodeDAO) CountAllEnabledAPINodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(APINodeStateEnabled).
		Count()
}

// CountAllEnabledAndOnAPINodes 计算启用中的API节点数量
func (this *APINodeDAO) CountAllEnabledAndOnAPINodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(APINodeStateEnabled).
		Attr("isOn", true).
		Count()
}

// CountAllEnabledAndOnOfflineAPINodes 计算API节点数量
func (this *APINodeDAO) CountAllEnabledAndOnOfflineAPINodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(APINodeStateEnabled).
		Attr("isOn", true).
		Where("(status IS NULL OR NOT JSON_EXTRACT(status, '$.isActive') OR UNIX_TIMESTAMP()-JSON_EXTRACT(status, '$.updatedAt')>60)").
		Count()
}

// ListEnabledAPINodes 列出单页的API节点
func (this *APINodeDAO) ListEnabledAPINodes(tx *dbs.Tx, offset int64, size int64) (result []*APINode, err error) {
	_, err = this.Query(tx).
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

// FindEnabledAPINodeIdWithAddr 根据主机名和端口获取ID
func (this *APINodeDAO) FindEnabledAPINodeIdWithAddr(tx *dbs.Tx, protocol string, host string, port int) (int64, error) {
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
		State(APINodeStateEnabled).
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
	return int64(one.(*APINode).Id), nil
}

// UpdateAPINodeStatus 设置API节点状态
func (this *APINodeDAO) UpdateAPINodeStatus(tx *dbs.Tx, apiNodeId int64, statusJSON []byte) error {
	_, err := this.Query(tx).
		Pk(apiNodeId).
		Set("status", statusJSON).
		Update()
	return err
}

// 生成唯一ID
func (this *APINodeDAO) genUniqueId(tx *dbs.Tx) (string, error) {
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

// CountAllLowerVersionNodes 计算所有节点中低于某个版本的节点数量
func (this *APINodeDAO) CountAllLowerVersionNodes(tx *dbs.Tx, version string) (int64, error) {
	return this.Query(tx).
		State(APINodeStateEnabled).
		Where("status IS NOT NULL").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("version", utils.VersionToLong(version)).
		Count()
}

// CountAllEnabledAPINodesWithSSLPolicyIds 计算使用SSL策略的所有API节点数量
func (this *APINodeDAO) CountAllEnabledAPINodesWithSSLPolicyIds(tx *dbs.Tx, sslPolicyIds []int64) (count int64, err error) {
	if len(sslPolicyIds) == 0 {
		return
	}
	policyStringIds := []string{}
	for _, policyId := range sslPolicyIds {
		policyStringIds = append(policyStringIds, strconv.FormatInt(policyId, 10))
	}
	return this.Query(tx).
		State(APINodeStateEnabled).
		Where("(FIND_IN_SET(JSON_EXTRACT(https, '$.sslPolicyRef.sslPolicyId'), :policyIds) OR FIND_IN_SET(JSON_EXTRACT(restHTTPS, '$.sslPolicyRef.sslPolicyId'), :policyIds))").
		Param("policyIds", strings.Join(policyStringIds, ",")).
		Count()
}
