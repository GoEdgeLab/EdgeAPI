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
	this.DAOObject.Init()
	this.DAOObject.OnUpdate(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
	this.DAOObject.OnInsert(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
	this.DAOObject.OnDelete(func() error {
		return SharedSysEventDAO.CreateEvent(NewServerChangeEvent())
	})
}

// 启用条目
func (this *HTTPFirewallRuleDAO) EnableHTTPFirewallRule(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallRuleDAO) DisableHTTPFirewallRule(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPFirewallRuleDAO) FindEnabledHTTPFirewallRule(id int64) (*HTTPFirewallRule, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPFirewallRuleStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallRule), err
}

// 组合配置
func (this *HTTPFirewallRuleDAO) ComposeFirewallRule(ruleId int64) (*firewallconfigs.HTTPFirewallRule, error) {
	rule, err := this.FindEnabledHTTPFirewallRule(ruleId)
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
func (this *HTTPFirewallRuleDAO) CreateOrUpdateRuleFromConfig(ruleConfig *firewallconfigs.HTTPFirewallRule) (int64, error) {
	op := NewHTTPFirewallRuleOperator()
	op.Id = ruleConfig.Id
	op.State = HTTPFirewallRuleStateEnabled
	op.IsOn = ruleConfig.IsOn
	op.Description = ruleConfig.Description
	op.Param = ruleConfig.Param
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
	_, err := this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}
