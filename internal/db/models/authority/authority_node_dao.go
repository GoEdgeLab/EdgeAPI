package authority

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
)

const (
	AuthorityNodeStateEnabled  = 1 // 已启用
	AuthorityNodeStateDisabled = 0 // 已禁用
)

type AuthorityNodeDAO dbs.DAO

func NewAuthorityNodeDAO() *AuthorityNodeDAO {
	return dbs.NewDAO(&AuthorityNodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeAuthorityNodes",
			Model:  new(AuthorityNode),
			PkName: "id",
		},
	}).(*AuthorityNodeDAO)
}

var SharedAuthorityNodeDAO *AuthorityNodeDAO

func init() {
	dbs.OnReady(func() {
		SharedAuthorityNodeDAO = NewAuthorityNodeDAO()
	})
}

// EnableAuthorityNode 启用条目
func (this *AuthorityNodeDAO) EnableAuthorityNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", AuthorityNodeStateEnabled).
		Update()
	return err
}

// DisableAuthorityNode 禁用条目
func (this *AuthorityNodeDAO) DisableAuthorityNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", AuthorityNodeStateDisabled).
		Update()
	return err
}

// FindEnabledAuthorityNode 查找启用中的条目
func (this *AuthorityNodeDAO) FindEnabledAuthorityNode(tx *dbs.Tx, id int64) (*AuthorityNode, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", AuthorityNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*AuthorityNode), err
}

// FindAuthorityNodeName 根据主键查找名称
func (this *AuthorityNodeDAO) FindAuthorityNodeName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindAllEnabledAuthorityNodes 列出所有可用认证节点
func (this *AuthorityNodeDAO) FindAllEnabledAuthorityNodes(tx *dbs.Tx) (result []*AuthorityNode, err error) {
	_, err = this.Query(tx).
		State(AuthorityNodeStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledAuthorityNodes 计算认证节点数量
func (this *AuthorityNodeDAO) CountAllEnabledAuthorityNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(AuthorityNodeStateEnabled).
		Count()
}

// ListEnabledAuthorityNodes 列出单页的认证节点
func (this *AuthorityNodeDAO) ListEnabledAuthorityNodes(tx *dbs.Tx, offset int64, size int64) (result []*AuthorityNode, err error) {
	_, err = this.Query(tx).
		State(AuthorityNodeStateEnabled).
		Offset(offset).
		Limit(size).
		Desc("order").
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CreateAuthorityNode 创建认证节点
func (this *AuthorityNodeDAO) CreateAuthorityNode(tx *dbs.Tx, name string, description string, isOn bool) (nodeId int64, err error) {
	uniqueId, err := this.GenUniqueId(tx)
	if err != nil {
		return 0, err
	}
	secret := rands.String(32)
	err = models.NewApiTokenDAO().CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleAuthority)
	if err != nil {
		return
	}

	var op = NewAuthorityNodeOperator()
	op.IsOn = isOn
	op.UniqueId = uniqueId
	op.Secret = secret
	op.Name = name
	op.Description = description
	op.State = AuthorityNodeStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return
	}

	return types.Int64(op.Id), nil
}

// UpdateAuthorityNode 修改认证节点
func (this *AuthorityNodeDAO) UpdateAuthorityNode(tx *dbs.Tx, nodeId int64, name string, description string, isOn bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	var op = NewAuthorityNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.Description = description
	op.IsOn = isOn
	err := this.Save(tx, op)
	return err
}

// FindEnabledAuthorityNodeWithUniqueId 根据唯一ID获取节点信息
func (this *AuthorityNodeDAO) FindEnabledAuthorityNodeWithUniqueId(tx *dbs.Tx, uniqueId string) (*AuthorityNode, error) {
	result, err := this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", AuthorityNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*AuthorityNode), err
}

// FindEnabledAuthorityNodeIdWithUniqueId 根据唯一ID获取节点ID
func (this *AuthorityNodeDAO) FindEnabledAuthorityNodeIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", AuthorityNodeStateEnabled).
		ResultPk().
		FindInt64Col(0)
}

// GenUniqueId 生成唯一ID
func (this *AuthorityNodeDAO) GenUniqueId(tx *dbs.Tx) (string, error) {
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
func (this *AuthorityNodeDAO) UpdateNodeStatus(tx *dbs.Tx, nodeId int64, nodeStatus *nodeconfigs.NodeStatus) error {
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
func (this *AuthorityNodeDAO) CountAllLowerVersionNodes(tx *dbs.Tx, version string) (int64, error) {
	return this.Query(tx).
		State(AuthorityNodeStateEnabled).
		Attr("isOn", true).
		Where("status IS NOT NULL").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("version", utils.VersionToLong(version)).
		Count()
}
