package models

import (
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
	MonitorNodeStateEnabled  = 1 // 已启用
	MonitorNodeStateDisabled = 0 // 已禁用
)

type MonitorNodeDAO dbs.DAO

func NewMonitorNodeDAO() *MonitorNodeDAO {
	return dbs.NewDAO(&MonitorNodeDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMonitorNodes",
			Model:  new(MonitorNode),
			PkName: "id",
		},
	}).(*MonitorNodeDAO)
}

var SharedMonitorNodeDAO *MonitorNodeDAO

func init() {
	dbs.OnReady(func() {
		SharedMonitorNodeDAO = NewMonitorNodeDAO()
	})
}

// EnableMonitorNode 启用条目
func (this *MonitorNodeDAO) EnableMonitorNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MonitorNodeStateEnabled).
		Update()
	return err
}

// DisableMonitorNode 禁用条目
func (this *MonitorNodeDAO) DisableMonitorNode(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MonitorNodeStateDisabled).
		Update()
	return err
}

// FindEnabledMonitorNode 查找启用中的条目
func (this *MonitorNodeDAO) FindEnabledMonitorNode(tx *dbs.Tx, id int64) (*MonitorNode, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", MonitorNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MonitorNode), err
}

// FindMonitorNodeName 根据主键查找名称
func (this *MonitorNodeDAO) FindMonitorNodeName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindAllEnabledMonitorNodes 列出所有可用监控节点
func (this *MonitorNodeDAO) FindAllEnabledMonitorNodes(tx *dbs.Tx) (result []*MonitorNode, err error) {
	_, err = this.Query(tx).
		State(MonitorNodeStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledMonitorNodes 计算监控节点数量
func (this *MonitorNodeDAO) CountAllEnabledMonitorNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(MonitorNodeStateEnabled).
		Count()
}

// ListEnabledMonitorNodes 列出单页的监控节点
func (this *MonitorNodeDAO) ListEnabledMonitorNodes(tx *dbs.Tx, offset int64, size int64) (result []*MonitorNode, err error) {
	_, err = this.Query(tx).
		State(MonitorNodeStateEnabled).
		Offset(offset).
		Limit(size).
		Desc("order").
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CreateMonitorNode 创建监控节点
func (this *MonitorNodeDAO) CreateMonitorNode(tx *dbs.Tx, name string, description string, isOn bool) (nodeId int64, err error) {
	uniqueId, err := this.GenUniqueId(tx)
	if err != nil {
		return 0, err
	}
	secret := rands.String(32)
	err = NewApiTokenDAO().CreateAPIToken(tx, uniqueId, secret, nodeconfigs.NodeRoleMonitor)
	if err != nil {
		return
	}

	op := NewMonitorNodeOperator()
	op.IsOn = isOn
	op.UniqueId = uniqueId
	op.Secret = secret
	op.Name = name
	op.Description = description
	op.State = NodeStateEnabled
	err = this.Save(tx, op)
	if err != nil {
		return
	}

	return types.Int64(op.Id), nil
}

// UpdateMonitorNode 修改监控节点
func (this *MonitorNodeDAO) UpdateMonitorNode(tx *dbs.Tx, nodeId int64, name string, description string, isOn bool) error {
	if nodeId <= 0 {
		return errors.New("invalid nodeId")
	}

	op := NewMonitorNodeOperator()
	op.Id = nodeId
	op.Name = name
	op.Description = description
	op.IsOn = isOn
	err := this.Save(tx, op)
	return err
}

// FindEnabledMonitorNodeWithUniqueId 根据唯一ID获取节点信息
func (this *MonitorNodeDAO) FindEnabledMonitorNodeWithUniqueId(tx *dbs.Tx, uniqueId string) (*MonitorNode, error) {
	result, err := this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", MonitorNodeStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MonitorNode), err
}

// FindEnabledMonitorNodeIdWithUniqueId 根据唯一ID获取节点ID
func (this *MonitorNodeDAO) FindEnabledMonitorNodeIdWithUniqueId(tx *dbs.Tx, uniqueId string) (int64, error) {
	return this.Query(tx).
		Attr("uniqueId", uniqueId).
		Attr("state", MonitorNodeStateEnabled).
		ResultPk().
		FindInt64Col(0)
}

// GenUniqueId 生成唯一ID
func (this *MonitorNodeDAO) GenUniqueId(tx *dbs.Tx) (string, error) {
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
func (this *MonitorNodeDAO) UpdateNodeStatus(tx *dbs.Tx, nodeId int64, statusJSON []byte) error {
	if statusJSON == nil {
		return nil
	}
	_, err := this.Query(tx).
		Pk(nodeId).
		Set("status", string(statusJSON)).
		Update()
	return err
}

// CountAllLowerVersionNodes 计算所有节点中低于某个版本的节点数量
func (this *MonitorNodeDAO) CountAllLowerVersionNodes(tx *dbs.Tx, version string) (int64, error) {
	return this.Query(tx).
		State(MonitorNodeStateEnabled).
		Where("status IS NOT NULL").
		Where("(JSON_EXTRACT(status, '$.buildVersionCode') IS NULL OR JSON_EXTRACT(status, '$.buildVersionCode')<:version)").
		Param("version", utils.VersionToLong(version)).
		Count()
}
