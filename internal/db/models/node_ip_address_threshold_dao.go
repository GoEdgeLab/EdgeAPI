package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

const (
	NodeIPAddressThresholdStateEnabled  = 1 // 已启用
	NodeIPAddressThresholdStateDisabled = 0 // 已禁用
)

type NodeIPAddressThresholdDAO dbs.DAO

func NewNodeIPAddressThresholdDAO() *NodeIPAddressThresholdDAO {
	return dbs.NewDAO(&NodeIPAddressThresholdDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeIPAddressThresholds",
			Model:  new(NodeIPAddressThreshold),
			PkName: "id",
		},
	}).(*NodeIPAddressThresholdDAO)
}

var SharedNodeIPAddressThresholdDAO *NodeIPAddressThresholdDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeIPAddressThresholdDAO = NewNodeIPAddressThresholdDAO()
	})
}

// EnableNodeIPAddressThreshold 启用条目
func (this *NodeIPAddressThresholdDAO) EnableNodeIPAddressThreshold(tx *dbs.Tx, id uint64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeIPAddressThresholdStateEnabled).
		Update()
	return err
}

// DisableNodeIPAddressThreshold 禁用条目
func (this *NodeIPAddressThresholdDAO) DisableNodeIPAddressThreshold(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeIPAddressThresholdStateDisabled).
		Update()
	return err
}

// FindEnabledNodeIPAddressThreshold 查找启用中的条目
func (this *NodeIPAddressThresholdDAO) FindEnabledNodeIPAddressThreshold(tx *dbs.Tx, id uint64) (*NodeIPAddressThreshold, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", NodeIPAddressThresholdStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeIPAddressThreshold), err
}

// FindAllEnabledThresholdsWithAddrId 查找所有阈值
func (this *NodeIPAddressThresholdDAO) FindAllEnabledThresholdsWithAddrId(tx *dbs.Tx, addrId int64) (result []*NodeIPAddressThreshold, err error) {
	_, err = this.Query(tx).
		Attr("addressId", addrId).
		State(NodeIPAddressThresholdStateEnabled).
		AscPk().
		Desc("order").
		Slice(&result).
		FindAll()
	if err != nil {
		return nil, err
	}

	// 过滤参数
	for _, threshold := range result {
		err := this.formatThreshold(tx, threshold)
		if err != nil {
			return nil, err
		}
	}

	return
}

// CountAllEnabledThresholdsWithAddrId 计算所有阈值数量
func (this *NodeIPAddressThresholdDAO) CountAllEnabledThresholdsWithAddrId(tx *dbs.Tx, addrId int64) (int64, error) {
	return this.Query(tx).
		Attr("addressId", addrId).
		State(NodeIPAddressThresholdStateEnabled).
		Count()
}

// FindThresholdNotifiedAt 查找上次通知时间
func (this *NodeIPAddressThresholdDAO) FindThresholdNotifiedAt(tx *dbs.Tx, thresholdId int64) (int64, error) {
	return this.Query(tx).
		Pk(thresholdId).
		Result("notifiedAt").
		FindInt64Col(0)
}

// UpdateThresholdNotifiedAt 设置上次通知时间
func (this *NodeIPAddressThresholdDAO) UpdateThresholdNotifiedAt(tx *dbs.Tx, thresholdId int64, timestamp int64) error {
	return this.Query(tx).
		Pk(thresholdId).
		Set("notifiedAt", timestamp).
		UpdateQuickly()
}

// CreateThreshold 创建阈值
func (this *NodeIPAddressThresholdDAO) CreateThreshold(tx *dbs.Tx, addressId int64, items []*nodeconfigs.NodeValueThresholdItemConfig, actions []*nodeconfigs.NodeValueThresholdActionConfig, order int) (int64, error) {
	if addressId <= 0 {
		return 0, errors.New("invalid addressId")
	}
	var op = NewNodeIPAddressThresholdOperator()
	op.Order = order
	op.AddressId = addressId

	if len(items) > 0 {
		itemsJSON, err := json.Marshal(items)
		if err != nil {
			return 0, err
		}
		op.Items = itemsJSON
	} else {
		op.Items = "[]"
	}

	if len(actions) > 0 {
		actionsJSON, err := json.Marshal(actions)
		if err != nil {
			return 0, err
		}
		op.Actions = actionsJSON
	} else {
		op.Actions = "[]"
	}

	op.State = NodeIPAddressThresholdStateEnabled
	return this.SaveInt64(tx, op)
}

// UpdateThreshold 修改阈值
func (this *NodeIPAddressThresholdDAO) UpdateThreshold(tx *dbs.Tx, thresholdId int64, items []*nodeconfigs.NodeValueThresholdItemConfig, actions []*nodeconfigs.NodeValueThresholdActionConfig, order int) error {
	if thresholdId <= 0 {
		return errors.New("invalid thresholdId")
	}
	var op = NewNodeIPAddressThresholdOperator()

	op.State = NodeIPAddressThresholdStateEnabled // 恢复状态
	if order >= 0 {
		op.Order = order
	}

	op.Id = thresholdId

	if len(items) > 0 {
		itemsJSON, err := json.Marshal(items)
		if err != nil {
			return err
		}
		op.Items = itemsJSON
	} else {
		op.Items = "[]"
	}

	if len(actions) > 0 {
		actionsJSON, err := json.Marshal(actions)
		if err != nil {
			return err
		}
		op.Actions = actionsJSON
	} else {
		op.Actions = "[]"
	}

	return this.Save(tx, op)
}

// DisableAllThresholdsWithAddrId 禁用所有阈值
func (this *NodeIPAddressThresholdDAO) DisableAllThresholdsWithAddrId(tx *dbs.Tx, addrId int64) error {
	return this.Query(tx).
		Attr("addressId", addrId).
		Set("state", NodeIPAddressThresholdStateDisabled).
		UpdateQuickly()
}

// 格式化阈值
func (this *NodeIPAddressThresholdDAO) formatThreshold(tx *dbs.Tx, threshold *NodeIPAddressThreshold) error {
	if len(threshold.Items) == 0 {
		return nil
	}
	var items = threshold.DecodeItems()
	for _, item := range items {
		if item.Item == nodeconfigs.IPAddressThresholdItemConnectivity {
			if item.Options == nil {
				continue
			}
			var groups = item.Options.GetSlice("groups")
			if len(groups) > 0 {
				var newGroups = []maps.Map{}
				for _, groupOne := range groups {
					var groupMap = maps.NewMap(groupOne)
					var groupId = groupMap.GetInt64("id")
					group, err := SharedReportNodeGroupDAO.FindEnabledReportNodeGroup(tx, groupId)
					if err != nil {
						return err
					}
					if group == nil {
						continue
					}
					newGroups = append(newGroups, maps.Map{
						"id":   group.Id,
						"name": group.Name,
					})
				}
				item.Options["groups"] = newGroups
			}
		}
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return err
	}
	threshold.Items = string(itemsJSON)

	return nil
}
