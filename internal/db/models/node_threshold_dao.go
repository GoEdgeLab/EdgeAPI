package models

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strings"
	"time"
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
func (this *NodeThresholdDAO) CreateThreshold(tx *dbs.Tx, role string, clusterId int64, nodeId int64, item nodeconfigs.NodeValueItem, param string, operator nodeconfigs.NodeValueOperator, valueJSON []byte, message string, sumMethod nodeconfigs.NodeValueSumMethod, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit, notifyDuration int32) (int64, error) {
	op := NewNodeThresholdOperator()
	op.Role = role
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
	op.NotifyDuration = notifyDuration
	op.IsOn = true
	op.State = NodeThresholdStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateThreshold 修改阈值
func (this *NodeThresholdDAO) UpdateThreshold(tx *dbs.Tx, thresholdId int64, item nodeconfigs.NodeValueItem, param string, operator nodeconfigs.NodeValueOperator, valueJSON []byte, message string, sumMethod nodeconfigs.NodeValueSumMethod, duration int32, durationUnit nodeconfigs.NodeValueDurationUnit, notifyDuration int32, isOn bool) error {
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
	op.NotifyDuration = notifyDuration
	op.IsOn = isOn
	return this.Save(tx, op)
}

// FindAllEnabledThresholds 列出所有阈值
func (this *NodeThresholdDAO) FindAllEnabledThresholds(tx *dbs.Tx, role string, clusterId int64, nodeId int64) (result []*NodeThreshold, err error) {
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
		Attr("role", role).
		Asc("IF(nodeId>0, 1, 0)").
		Desc("order").
		AscPk().
		FindAll()
	return
}

// FindAllEnabledAndOnClusterThresholds 查询集群专属的阈值设置
func (this *NodeThresholdDAO) FindAllEnabledAndOnClusterThresholds(tx *dbs.Tx, role string, clusterId int64, item string) (result []*NodeThreshold, err error) {
	if clusterId <= 0 {
		return
	}
	_, err = this.Query(tx).
		Attr("role", role).
		Attr("clusterId", clusterId).
		Attr("nodeId", 0).
		Attr("item", item).
		Attr("isOn", true).
		State(NodeThresholdStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// FindAllEnabledAndOnNodeThresholds 查询节点专属的阈值设置
func (this *NodeThresholdDAO) FindAllEnabledAndOnNodeThresholds(tx *dbs.Tx, role string, nodeId int64, item string) (result []*NodeThreshold, err error) {
	if nodeId <= 0 {
		return
	}
	_, err = this.Query(tx).
		Attr("role", role).
		Attr("nodeId", nodeId).
		Attr("item", item).
		Attr("isOn", true).
		State(NodeThresholdStateEnabled).
		Desc("order").
		AscPk().
		Slice(&result).
		FindAll()
	return
}

// CountAllEnabledThresholds 计算阈值的数量
func (this *NodeThresholdDAO) CountAllEnabledThresholds(tx *dbs.Tx, role string, clusterId int64, nodeId int64) (int64, error) {
	if clusterId <= 0 && nodeId <= 0 {
		return 0, nil
	}
	query := this.Query(tx)
	query.Attr("role", role)
	if clusterId > 0 {
		query.Attr("clusterId", clusterId)
	}
	if nodeId > 0 {
		query.Attr("nodeId", nodeId)
	}
	query.State(NodeThresholdStateEnabled)
	return query.Count()
}

// FireNodeThreshold 触发相关阈值设置
func (this *NodeThresholdDAO) FireNodeThreshold(tx *dbs.Tx, role string, nodeId int64, item string) error {
	clusterId, err := SharedNodeDAO.FindNodeClusterId(tx, nodeId)
	if err != nil {
		return err
	}
	if clusterId == 0 {
		return nil
	}

	// 集群相关阈值
	var thresholds []*NodeThreshold
	{
		clusterThresholds, err := this.FindAllEnabledAndOnClusterThresholds(tx, role, clusterId, item)
		if err != nil {
			return err
		}
		thresholds = append(thresholds, clusterThresholds...)
	}

	// 节点相关阈值
	{
		nodeThresholds, err := this.FindAllEnabledAndOnNodeThresholds(tx, role, nodeId, item)
		if err != nil {
			return err
		}
		thresholds = append(thresholds, nodeThresholds...)
	}

	if len(thresholds) > 0 {
		for _, threshold := range thresholds {
			if len(threshold.Param) == 0 || threshold.Duration <= 0 {
				continue
			}
			paramValue, err := SharedNodeValueDAO.SumValues(tx, role, nodeId, item, threshold.Param, threshold.SumMethod, types.Int32(threshold.Duration), threshold.DurationUnit)
			if err != nil {
				return err
			}
			originValue := nodeconfigs.UnmarshalNodeValue([]byte(threshold.Value))
			thresholdValue := types.Float64(originValue)
			isMatched := nodeconfigs.CompareNodeValue(threshold.Operator, paramValue, thresholdValue)
			if isMatched {
				// TODO 执行其他动作

				// 是否已经通知过
				if threshold.NotifyDuration > 0 && threshold.NotifiedAt > 0 && time.Now().Unix()-int64(threshold.NotifiedAt) < int64(threshold.NotifyDuration*60) {
					continue
				}

				// 创建消息
				nodeName, err := SharedNodeDAO.FindNodeName(tx, nodeId)
				if err != nil {
					return err
				}
				itemName := nodeconfigs.FindNodeValueItemName(threshold.Item)
				paramName := nodeconfigs.FindNodeValueItemParamName(threshold.Item, threshold.Param)
				operatorName := nodeconfigs.FindNodeValueOperatorName(threshold.Operator)

				subject := "节点 \"" + nodeName + "\" " + itemName + " 达到阈值"
				body := "节点 \"" + nodeName + "\" " + itemName + " 达到阈值\n阈值设置：" + paramName + " " + operatorName + " " + originValue + "\n当前值：" + fmt.Sprintf("%.2f", paramValue) + "\n触发时间：" + timeutil.Format("Y-m-d H:i:s")
				if len(threshold.Message) > 0 {
					body = threshold.Message
					body = strings.Replace(body, "${item.name}", itemName, -1)
					body = strings.Replace(body, "${value}", fmt.Sprintf("%.2f", paramValue), -1)
				}
				err = SharedMessageDAO.CreateNodeMessage(tx, clusterId, nodeId, MessageTypeThresholdSatisfied, MessageLevelWarning, subject, body, maps.Map{}.AsJSON())
				if err != nil {
					return err
				}

				// 设置通知时间
				_, err = this.Query(tx).
					Pk(threshold.Id).
					Set("notifiedAt", time.Now().Unix()).
					Update()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
