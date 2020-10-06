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
	HTTPFirewallRuleGroupStateEnabled  = 1 // 已启用
	HTTPFirewallRuleGroupStateDisabled = 0 // 已禁用
)

type HTTPFirewallRuleGroupDAO dbs.DAO

func NewHTTPFirewallRuleGroupDAO() *HTTPFirewallRuleGroupDAO {
	return dbs.NewDAO(&HTTPFirewallRuleGroupDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeHTTPFirewallRuleGroups",
			Model:  new(HTTPFirewallRuleGroup),
			PkName: "id",
		},
	}).(*HTTPFirewallRuleGroupDAO)
}

var SharedHTTPFirewallRuleGroupDAO = NewHTTPFirewallRuleGroupDAO()

// 初始化
func (this *HTTPFirewallRuleGroupDAO) Init() {
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
func (this *HTTPFirewallRuleGroupDAO) EnableHTTPFirewallRuleGroup(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleGroupStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *HTTPFirewallRuleGroupDAO) DisableHTTPFirewallRuleGroup(id int64) error {
	_, err := this.Query().
		Pk(id).
		Set("state", HTTPFirewallRuleGroupStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *HTTPFirewallRuleGroupDAO) FindEnabledHTTPFirewallRuleGroup(id int64) (*HTTPFirewallRuleGroup, error) {
	result, err := this.Query().
		Pk(id).
		Attr("state", HTTPFirewallRuleGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallRuleGroup), err
}

// 根据主键查找名称
func (this *HTTPFirewallRuleGroupDAO) FindHTTPFirewallRuleGroupName(id int64) (string, error) {
	return this.Query().
		Pk(id).
		Result("name").
		FindStringCol("")
}

// 组合配置
func (this *HTTPFirewallRuleGroupDAO) ComposeFirewallRuleGroup(groupId int64) (*firewallconfigs.HTTPFirewallRuleGroup, error) {
	group, err := this.FindEnabledHTTPFirewallRuleGroup(groupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}
	config := &firewallconfigs.HTTPFirewallRuleGroup{}
	config.Id = int64(group.Id)
	config.IsOn = group.IsOn == 1
	config.Name = group.Name
	config.Description = group.Description
	config.Code = group.Code

	if IsNotNull(group.Sets) {
		setRefs := []*firewallconfigs.HTTPFirewallRuleSetRef{}
		err = json.Unmarshal([]byte(group.Sets), &setRefs)
		if err != nil {
			return nil, err
		}
		for _, setRef := range setRefs {
			setConfig, err := SharedHTTPFirewallRuleSetDAO.ComposeFirewallRuleSet(setRef.SetId)
			if err != nil {
				return nil, err
			}
			if setConfig != nil {
				config.SetRefs = append(config.SetRefs, setRef)
				config.Sets = append(config.Sets, setConfig)
			}
		}
	}

	return config, nil
}

// 从配置中创建分组
func (this *HTTPFirewallRuleGroupDAO) CreateGroupFromConfig(groupConfig *firewallconfigs.HTTPFirewallRuleGroup) (int64, error) {
	op := NewHTTPFirewallRuleGroupOperator()
	op.IsOn = groupConfig.IsOn
	op.Name = groupConfig.Name
	op.Description = groupConfig.Description
	op.State = HTTPFirewallRuleGroupStateEnabled
	op.Code = groupConfig.Code

	// sets
	setRefs := []*firewallconfigs.HTTPFirewallRuleSetRef{}
	for _, setConfig := range groupConfig.Sets {
		setId, err := SharedHTTPFirewallRuleSetDAO.CreateSetFromConfig(setConfig)
		if err != nil {
			return 0, err
		}
		setRefs = append(setRefs, &firewallconfigs.HTTPFirewallRuleSetRef{
			IsOn:  true,
			SetId: setId,
		})
	}
	setRefsJSON, err := json.Marshal(setRefs)
	if err != nil {
		return 0, err
	}
	op.Sets = setRefsJSON
	_, err = this.Save(op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// 修改开启状态
func (this *HTTPFirewallRuleGroupDAO) UpdateGroupIsOn(groupId int64, isOn bool) error {
	_, err := this.Query().
		Pk(groupId).
		Set("isOn", isOn).
		Update()
	return err
}
