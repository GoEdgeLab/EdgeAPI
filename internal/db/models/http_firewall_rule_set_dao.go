package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPFirewallRuleSetStateEnabled  = 1 // 已启用
	HTTPFirewallRuleSetStateDisabled = 0 // 已禁用
)

type HTTPFirewallRuleSetDAO dbs.DAO

func NewHTTPFirewallRuleSetDAO() *HTTPFirewallRuleSetDAO {
	return dbs.NewDAO(&HTTPFirewallRuleSetDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFirewallRuleSets",
			Model:  new(HTTPFirewallRuleSet),
			PkName: "id",
		},
	}).(*HTTPFirewallRuleSetDAO)
}

var SharedHTTPFirewallRuleSetDAO = NewHTTPFirewallRuleSetDAO()

// 初始化
func (this *HTTPFirewallRuleSetDAO) Init() {
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
func (this *HTTPFirewallRuleSetDAO) EnableHTTPFirewallRuleSet(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleSetStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallRuleSetDAO) DisableHTTPFirewallRuleSet(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleSetStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPFirewallRuleSetDAO) FindEnabledHTTPFirewallRuleSet(id int64) (*HTTPFirewallRuleSet, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPFirewallRuleSetStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallRuleSet), err
}

// 根据主键查找名称
func (this *HTTPFirewallRuleSetDAO) FindHTTPFirewallRuleSetName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 组合配置
func (this *HTTPFirewallRuleSetDAO) ComposeFirewallRuleSet(setId int64) (*firewallconfigs.HTTPFirewallRuleSet, error) {
	set, err := this.FindEnabledHTTPFirewallRuleSet(setId)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, nil
	}
	config := &firewallconfigs.HTTPFirewallRuleSet{}
	config.Id = int64(set.Id)
	config.IsOn = set.IsOn == 1
	config.Name = set.Name
	config.Description = set.Description
	config.Code = set.Code
	config.Connector = set.Connector

	if IsNotNull(set.Rules) {
		ruleRefs := []*firewallconfigs.HTTPFirewallRuleRef{}
		err = json.Unmarshal([]byte(set.Rules), &ruleRefs)
		if err != nil {
			return nil, err
		}
		for _, ruleRef := range ruleRefs {
			ruleConfig, err := SharedHTTPFirewallRuleDAO.ComposeFirewallRule(ruleRef.RuleId)
			if err != nil {
				return nil, err
			}
			if ruleConfig != nil {
				config.RuleRefs = append(config.RuleRefs, ruleRef)
				config.Rules = append(config.Rules, ruleConfig)
			}
		}
	}

	config.Action = set.Action
	if IsNotNull(set.ActionOptions) {
		options := maps.Map{}
		err = json.Unmarshal([]byte(set.ActionOptions), &options)
		if err != nil {
			return nil, err
		}
		config.ActionOptions = options
	}

	return config, nil
}

// 从配置中创建规则集
func (this *HTTPFirewallRuleSetDAO) CreateSetFromConfig(setConfig *firewallconfigs.HTTPFirewallRuleSet) (int64, error) {
	op := NewHTTPFirewallRuleSetOperator()
	op.State = HTTPFirewallRuleSetStateEnabled
	op.IsOn = setConfig.IsOn
	op.Name = setConfig.Name
	op.Description = setConfig.Description
	op.Connector = setConfig.Connector
	op.Action = setConfig.Action
	op.Code = setConfig.Code

	if setConfig.ActionOptions != nil {
		actionOptionsJSON, err := json.Marshal(setConfig.ActionOptions)
		if err != nil {
			return 0, err
		}
		op.ActionOptions = actionOptionsJSON
	}

	// rules
	ruleRefs := []*firewallconfigs.HTTPFirewallRuleRef{}
	for _, ruleConfig := range setConfig.Rules {
		ruleId, err := SharedHTTPFirewallRuleDAO.CreateRuleFromConfig(ruleConfig)
		if err != nil {
			return 0, err
		}
		ruleRefs = append(ruleRefs, &firewallconfigs.HTTPFirewallRuleRef{
			IsOn:   true,
			RuleId: ruleId,
		})
	}
	ruleRefsJSON, err := json.Marshal(ruleRefs)
	if err != nil {
		return 0, err
	}
	op.Rules = ruleRefsJSON
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}
