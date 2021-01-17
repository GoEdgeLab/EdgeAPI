package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPFirewallRuleStateEnabled  = 1 // 已启用
	HTTPFirewallRuleStateDisabled = 0 // 已禁用
)

type HTTPFirewallRuleDAO dbs.DAO

func NewHTTPFirewallRuleDAO() *HTTPFirewallRuleDAO {
	return dbs.NewDAO(&HTTPFirewallRuleDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFirewallRules",
			Model:  new(HTTPFirewallRule),
			PkName: "id",
		},
	}).(*HTTPFirewallRuleDAO)
}

var SharedHTTPFirewallRuleDAO *HTTPFirewallRuleDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPFirewallRuleDAO = NewHTTPFirewallRuleDAO()
	})
}

// 初始化
func (this *HTTPFirewallRuleDAO) Init() {
	_ = this.DAOObject.Init()
}

// 启用条目
func (this *HTTPFirewallRuleDAO) EnableHTTPFirewallRule(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPFirewallRuleStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallRuleDAO) DisableHTTPFirewallRule(tx *dbs.Tx, ruleId int64) error {
	_, err := this.Query(tx).
		Pk(ruleId).
		Set("state", HTTPFirewallRuleStateDisabled).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, ruleId)
}

// 查找启用中的条目
func (this *HTTPFirewallRuleDAO) FindEnabledHTTPFirewallRule(tx *dbs.Tx, id int64) (*HTTPFirewallRule, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPFirewallRuleStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallRule), err
}

// 组合配置
func (this *HTTPFirewallRuleDAO) ComposeFirewallRule(tx *dbs.Tx, ruleId int64) (*firewallconfigs.HTTPFirewallRule, error) {
	rule, err := this.FindEnabledHTTPFirewallRule(tx, ruleId)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, nil
	}
	config := &firewallconfigs.HTTPFirewallRule{}
	config.Id = int64(rule.Id)
	config.IsOn = rule.IsOn == 1
	config.Param = rule.Param

	paramFilters := []*firewallconfigs.ParamFilter{}
	if IsNotNull(rule.ParamFilters) {
		err = json.Unmarshal([]byte(rule.ParamFilters), &paramFilters)
		if err != nil {
			return nil, err
		}
	}
	config.ParamFilters = paramFilters

	config.Operator = rule.Operator
	config.Value = rule.Value
	config.IsCaseInsensitive = rule.IsCaseInsensitive == 1

	if IsNotNull(rule.CheckpointOptions) {
		checkpointOptions := map[string]interface{}{}
		err = json.Unmarshal([]byte(rule.CheckpointOptions), &checkpointOptions)
		if err != nil {
			return nil, err
		}
		config.CheckpointOptions = checkpointOptions
	}

	config.Description = rule.Description

	return config, nil
}

// 从配置中配置规则
func (this *HTTPFirewallRuleDAO) CreateOrUpdateRuleFromConfig(tx *dbs.Tx, ruleConfig *firewallconfigs.HTTPFirewallRule) (int64, error) {
	op := NewHTTPFirewallRuleOperator()
	op.Id = ruleConfig.Id
	op.State = HTTPFirewallRuleStateEnabled
	op.IsOn = ruleConfig.IsOn
	op.Description = ruleConfig.Description
	op.Param = ruleConfig.Param

	if len(ruleConfig.ParamFilters) == 0 {
		op.ParamFilters = "[]"
	} else {
		paramFilters, err := json.Marshal(ruleConfig.ParamFilters)
		if err != nil {
			return 0, err
		}
		op.ParamFilters = paramFilters
	}

	op.Value = ruleConfig.Value
	op.IsCaseInsensitive = ruleConfig.IsCaseInsensitive
	op.Operator = ruleConfig.Operator

	if ruleConfig.CheckpointOptions != nil {
		checkpointOptionsJSON, err := json.Marshal(ruleConfig.CheckpointOptions)
		if err != nil {
			return 0, err
		}
		op.CheckpointOptions = checkpointOptionsJSON
	}
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}

	// 通知更新
	if ruleConfig.Id > 0 {
		err := this.NotifyUpdate(tx, ruleConfig.Id)
		if err != nil {
			return 0, err
		}
	}

	return types.Int64(op.Id), nil
}

// 通知更新
func (this *HTTPFirewallRuleDAO) NotifyUpdate(tx *dbs.Tx, ruleId int64) error {
	setId, err := SharedHTTPFirewallRuleSetDAO.FindEnabledRuleSetIdWithRuleId(tx, ruleId)
	if err != nil {
		return err
	}
	if setId > 0 {
		return SharedHTTPFirewallRuleSetDAO.NotifyUpdate(tx, setId)
	}
	return nil
}
