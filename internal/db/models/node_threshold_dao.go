package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
)

const (
	NodeThresholdStateEnabled  = 1 // 已启用
	NodeThresholdStateDisabled = 0 // 已禁用
)

type NodeThresholdDAO dbs.DAO

func NewNodeThresholdDAO() *NodeThresholdDAO {
	return dbs.NewDAO(&NodeThresholdDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeThresholds",
			Model:  new(NodeThreshold),
			PkName: "id",
		},
	}).(*NodeThresholdDAO)
}

var SharedNodeThresholdDAO *NodeThresholdDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeThresholdDAO = NewNodeThresholdDAO()
	})
}

// EnableNodeThreshold 启用条目
func (this *NodeThresholdDAO) EnableNodeThreshold(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeThresholdStateEnabled).
		Update()
	return err
}

// DisableNodeThreshold 禁用条目
func (this *NodeThresholdDAO) DisableNodeThreshold(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeThresholdStateDisabled).
		Update()
	return err
}

// FindEnabledNodeThreshold 查找启用中的条目
func (this *NodeThresholdDAO) FindEnabledNodeThreshold(tx *dbs.Tx, id int64) (*NodeThreshold, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeThresholdStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeThreshold), err
}

// CreateThreshold 创建阈值
func (this *NodeThresholdDAO) CreateThreshold(tx *dbs.Tx, clusterId int64, nodeId int64, item nodeconfigs.NodeValueItem, param string, operator nodeconfigs.NodeValueOperator, valueJSON []byte, message string, sumMethod nodeconfigs.NodeValueSumMethod, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit) (int64, error) {
	op := NewNodeThresholdOperator()
	op.ClusterId = clusterId
	op.NodeId = nodeId
	op.Item = item
	op.Param = param
	op.Operator = operator
	op.Value = valueJSON
	op.Message = message
	op.SumMethod = sumMethod
	op.Duration = duration
	op.DurationUnit = durationUnit
	op.IsOn = true
	op.State = NodeThresholdStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateThreshold 修改阈值
func (this *NodeThresholdDAO) UpdateThreshold(tx *dbs.Tx, thresholdId int64, item nodeconfigs.NodeValueItem, param string, operator nodeconfigs.NodeValueOperator, valueJSON []byte, message string, sumMethod nodeconfigs.NodeValueSumMethod, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit, isOn bool) error {
	if thresholdId <= 0 {
		return errors.New("invalid thresholdId")
	}
	op := NewNodeThresholdOperator()
	op.Id = thresholdId
	op.Item = item
	op.Param = param
	op.Operator = operator
	op.Value = valueJSON
	op.Message = message
	op.SumMethod = sumMethod
	op.Duration = duration
	op.DurationUnit = durationUnit
	op.IsOn = isOn
	return this.Save(tx, op)
}

// FindAllEnabledThresholds 列出所有阈值
func (this *NodeThresholdDAO) FindAllEnabledThresholds(tx *dbs.Tx, clusterId int64, nodeId int64) (result []*NodeThreshold, err error) {
	if clusterId <= 0 && nodeId <= 0 {
		return
	}
	query := this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if nodeId > 0 {
		query.Attr("nodeId", nodeId)
	}
	query.State(NodeThresholdStateEnabled)
	query.Slice(&result)
	_, err = query.
		AscPk().
		FindAll()
	return
}

// CountAllEnabledThresholds 计算阈值的数量
func (this *NodeThresholdDAO) CountAllEnabledThresholds(tx *dbs.Tx, clusterId int64, nodeId int64) (int64, error) {
	if clusterId <= 0 && nodeId <= 0 {
		return 0, nil
	}
	query := this.Query(tx)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if nodeId > 0 {
		query.Attr("nodeId", nodeId)
	}
	query.State(NodeThresholdStateEnabled)
	return query.Count()
}
