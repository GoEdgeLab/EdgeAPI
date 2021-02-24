package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

const (
	NodeClusterFirewallActionStateEnabled  = 1 // 已启用
	NodeClusterFirewallActionStateDisabled = 0 // 已禁用
)

type NodeClusterFirewallActionDAO dbs.DAO

func NewNodeClusterFirewallActionDAO() *NodeClusterFirewallActionDAO {
	return dbs.NewDAO(&NodeClusterFirewallActionDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeNodeClusterFirewallActions",
			Model:  new(NodeClusterFirewallAction),
			PkName: "id",
		},
	}).(*NodeClusterFirewallActionDAO)
}

var SharedNodeClusterFirewallActionDAO *NodeClusterFirewallActionDAO

func init() {
	dbs.OnReady(func() {
		SharedNodeClusterFirewallActionDAO = NewNodeClusterFirewallActionDAO()
	})
}

// 启用条目
func (this *NodeClusterFirewallActionDAO) EnableFirewallAction(tx *dbs.Tx, id uint32) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", NodeClusterFirewallActionStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *NodeClusterFirewallActionDAO) DisableFirewallAction(tx *dbs.Tx, actionId int64) error {
	_, err := this.Query(tx).
		Pk(actionId).
		Set("state", NodeClusterFirewallActionStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, actionId)
}

// 查找启用中的条目
func (this *NodeClusterFirewallActionDAO) FindEnabledFirewallAction(tx *dbs.Tx, actionId int64) (*NodeClusterFirewallAction, error) {
	result, err := this.Query(tx).
		Pk(actionId).
		Attr("state", NodeClusterFirewallActionStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*NodeClusterFirewallAction), err
}

// 根据主键查找名称
func (this *NodeClusterFirewallActionDAO) FindFirewallActionName(tx *dbs.Tx, id uint32) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 创建动作
func (this *NodeClusterFirewallActionDAO) CreateFirewallAction(tx *dbs.Tx, adminId int64, clusterId int64, name string, eventLevel, actionType firewallconfigs.FirewallActionType, params maps.Map) (int64, error) {
	if params == nil {
		params = maps.Map{}
	}

	op := NewNodeClusterFirewallActionOperator()
	op.AdminId = adminId
	op.ClusterId = clusterId
	op.Name = name
	op.EventLevel = eventLevel
	op.Type = actionType
	op.Params = params.AsJSON()
	op.State = NodeClusterFirewallActionStateEnabled
	actionId, err := this.SaveInt64(tx, op)
	if err != nil {
		return 0, err
	}
	err = this.NotifyUpdate(tx, actionId)
	if err != nil {
		return 0, err
	}
	return actionId, nil
}

// 修改动作
func (this *NodeClusterFirewallActionDAO) UpdateFirewallAction(tx *dbs.Tx, actionId int64, name string, eventLevel string, actionType firewallconfigs.FirewallActionType, params maps.Map) error {
	if actionId <= 0 {
		return errors.New("invalid actionId")
	}

	if params == nil {
		params = maps.Map{}
	}

	op := NewNodeClusterFirewallActionOperator()
	op.Id = actionId
	op.Name = name
	op.EventLevel = eventLevel
	op.Type = actionType
	op.Params = params.AsJSON()
	_, err := this.SaveInt64(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, actionId)
}

// 查找所有集群的动作
func (this *NodeClusterFirewallActionDAO) FindAllEnabledFirewallActions(tx *dbs.Tx, clusterId int64) (result []*NodeClusterFirewallAction, err error) {
	_, err = this.Query(tx).
		Attr("clusterId", clusterId).
		State(NodeClusterFirewallActionStateEnabled).
		Slice(&result).
		FindAll()
	return
}

// 组合配置
func (this *NodeClusterFirewallActionDAO) ComposeFirewallActionConfig(tx *dbs.Tx, action *NodeClusterFirewallAction) (*firewallconfigs.FirewallActionConfig, error) {
	if action == nil {
		return nil, nil
	}
	config := &firewallconfigs.FirewallActionConfig{}
	config.Id = int64(action.Id)
	config.Type = action.Type
	config.EventLevel = action.EventLevel

	params, err := action.DecodeParams()
	if err != nil {
		return nil, err
	}
	config.Params = params

	return config, nil
}

// 计算动作数量
func (this *NodeClusterFirewallActionDAO) CountAllEnabledFirewallActions(tx *dbs.Tx, clusterId int64) (int64, error) {
	return this.Query(tx).
		State(NodeClusterFirewallActionStateEnabled).
		Attr("clusterId", clusterId).
		Count()
}

// 通知更新
func (this *NodeClusterFirewallActionDAO) NotifyUpdate(tx *dbs.Tx, actionId int64) error {
	clusterId, err := this.Query(tx).
		Pk(actionId).
		Result("clusterId").
		FindInt64Col(0)
	if err != nil {
		return err
	}
	if clusterId > 0 {
		return SharedNodeClusterDAO.NotifyUpdate(tx, clusterId)
	}
	return nil
}
