package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

const (
	HTTPFirewallPolicyStateEnabled  = 1 // 已启用
	HTTPFirewallPolicyStateDisabled = 0 // 已禁用
)

type HTTPFirewallPolicyDAO dbs.DAO

func NewHTTPFirewallPolicyDAO() *HTTPFirewallPolicyDAO {
	return dbs.NewDAO(&HTTPFirewallPolicyDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFirewallPolicies",
			Model:  new(HTTPFirewallPolicy),
			PkName: "id",
		},
	}).(*HTTPFirewallPolicyDAO)
}

var SharedHTTPFirewallPolicyDAO = NewHTTPFirewallPolicyDAO()

// 初始化
func (this *HTTPFirewallPolicyDAO) Init() {
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
func (this *HTTPFirewallPolicyDAO) EnableHTTPFirewallPolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallPolicyStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallPolicyDAO) DisableHTTPFirewallPolicy(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallPolicyStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPFirewallPolicyDAO) FindEnabledHTTPFirewallPolicy(id int64) (*HTTPFirewallPolicy, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPFirewallPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallPolicy), err
}

// 根据主键查找名称
func (this *HTTPFirewallPolicyDAO) FindHTTPFirewallPolicyName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 查找所有可用策略
func (this *HTTPFirewallPolicyDAO) FindAllEnabledFirewallPolicies() (result []*HTTPFirewallPolicy, err error) {
	_, err = this.Query().
		State(HTTPFirewallPolicyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 创建策略
func (this *HTTPFirewallPolicyDAO) CreateFirewallPolicy(isOn bool, name string, description string, inboundJSON []byte, outboundJSON []byte) (int64, error) {
	op := NewHTTPFirewallPolicyOperator()
	op.State = HTTPFirewallPolicyStateEnabled
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	if len(inboundJSON) > 0 {
		op.Inbound = inboundJSON
	}
	if len(outboundJSON) > 0 {
		op.Outbound = outboundJSON
	}
	_, err := this.Save(op)
	return types.Int64(op.Id), err
}

// 修改策略的Inbound和Outbound
func (this *HTTPFirewallPolicyDAO) UpdateFirewallPolicyInboundAndOutbound(policyId int64, inboundJSON []byte, outboundJSON []byte) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}
	op := NewHTTPFirewallPolicyOperator()
	op.Id = policyId
	if len(inboundJSON) > 0 {
		op.Inbound = inboundJSON
	} else {
		op.Inbound = "null"
	}
	if len(outboundJSON) > 0 {
		op.Outbound = outboundJSON
	} else {
		op.Outbound = "null"
	}
	_, err := this.Save(op)
	return err
}

// 修改策略
func (this *HTTPFirewallPolicyDAO) UpdateFirewallPolicy(policyId int64, isOn bool, name string, description string, inboundJSON []byte, outboundJSON []byte) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}
	op := NewHTTPFirewallPolicyOperator()
	op.Id = policyId
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	if len(inboundJSON) > 0 {
		op.Inbound = inboundJSON
	} else {
		op.Inbound = "null"
	}
	if len(outboundJSON) > 0 {
		op.Outbound = outboundJSON
	} else {
		op.Outbound = "null"
	}
	_, err := this.Save(op)
	return err
}

// 计算所有可用的策略数量
func (this *HTTPFirewallPolicyDAO) CountAllEnabledFirewallPolicies() (int64, error) {
	return this.Query().
		State(HTTPFirewallPolicyStateEnabled).
		Count()
}

// 列出单页的策略
func (this *HTTPFirewallPolicyDAO) ListEnabledFirewallPolicies(offset int64, size int64) (result []*HTTPFirewallPolicy, err error) {
	_, err = this.Query().
		State(HTTPFirewallPolicyStateEnabled).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// 组合策略配置
func (this *HTTPFirewallPolicyDAO) ComposeFirewallPolicy(policyId int64) (*firewallconfigs.HTTPFirewallPolicy, error) {
	policy, err := this.FindEnabledHTTPFirewallPolicy(policyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}

	config := &firewallconfigs.HTTPFirewallPolicy{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn == 1
	config.Name = policy.Name
	config.Description = policy.Description

	// Inbound
	if IsNotNull(policy.Inbound) {
		inbound := &firewallconfigs.HTTPFirewallInboundConfig{}
		err = json.Unmarshal([]byte(policy.Inbound), inbound)
		if err != nil {
			return nil, err
		}
		if len(inbound.GroupRefs) > 0 {
			resultGroupRefs := []*firewallconfigs.HTTPFirewallRuleGroupRef{}
			resultGroups := []*firewallconfigs.HTTPFirewallRuleGroup{}

			for _, groupRef := range inbound.GroupRefs {
				groupConfig, err := SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(groupRef.GroupId)
				if err != nil {
					return nil, err
				}
				if groupConfig != nil {
					resultGroupRefs = append(resultGroupRefs, groupRef)
					resultGroups = append(resultGroups, groupConfig)
				}
			}

			inbound.GroupRefs = resultGroupRefs
			inbound.Groups = resultGroups
		}
		config.Inbound = inbound
	}

	// Outbound
	if IsNotNull(policy.Outbound) {
		outbound := &firewallconfigs.HTTPFirewallOutboundConfig{}
		err = json.Unmarshal([]byte(policy.Outbound), outbound)
		if err != nil {
			return nil, err
		}
		if len(outbound.GroupRefs) > 0 {
			resultGroupRefs := []*firewallconfigs.HTTPFirewallRuleGroupRef{}
			resultGroups := []*firewallconfigs.HTTPFirewallRuleGroup{}

			for _, groupRef := range outbound.GroupRefs {
				groupConfig, err := SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(groupRef.GroupId)
				if err != nil {
					return nil, err
				}
				if groupConfig != nil {
					resultGroupRefs = append(resultGroupRefs, groupRef)
					resultGroups = append(resultGroups, groupConfig)
				}
			}

			outbound.GroupRefs = resultGroupRefs
			outbound.Groups = resultGroups
		}
		config.Outbound = outbound
	}

	return config, nil
}
