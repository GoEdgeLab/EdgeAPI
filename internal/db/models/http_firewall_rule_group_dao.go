package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
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

var SharedHTTPFirewallRuleGroupDAO *HTTPFirewallRuleGroupDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPFirewallRuleGroupDAO = NewHTTPFirewallRuleGroupDAO()
	})
}

// Init 初始化
func (this *HTTPFirewallRuleGroupDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPFirewallRuleGroup 启用条目
func (this *HTTPFirewallRuleGroupDAO) EnableHTTPFirewallRuleGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPFirewallRuleGroupStateEnabled).
		Update()
	return err
}

// DisableHTTPFirewallRuleGroup 禁用条目
func (this *HTTPFirewallRuleGroupDAO) DisableHTTPFirewallRuleGroup(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPFirewallRuleGroupStateDisabled).
		Update()
	return err
}

// FindEnabledHTTPFirewallRuleGroup 查找启用中的条目
func (this *HTTPFirewallRuleGroupDAO) FindEnabledHTTPFirewallRuleGroup(tx *dbs.Tx, id int64) (*HTTPFirewallRuleGroup, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPFirewallRuleGroupStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallRuleGroup), err
}

// FindHTTPFirewallRuleGroupName 根据主键查找名称
func (this *HTTPFirewallRuleGroupDAO) FindHTTPFirewallRuleGroupName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// ComposeFirewallRuleGroup 组合配置
func (this *HTTPFirewallRuleGroupDAO) ComposeFirewallRuleGroup(tx *dbs.Tx, groupId int64) (*firewallconfigs.HTTPFirewallRuleGroup, error) {
	group, err := this.FindEnabledHTTPFirewallRuleGroup(tx, groupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}
	config := &firewallconfigs.HTTPFirewallRuleGroup{}
	config.Id = int64(group.Id)
	config.IsOn = group.IsOn
	config.Name = group.Name
	config.Description = group.Description
	config.Code = group.Code
	config.IsTemplate = group.IsTemplate

	if IsNotNull(group.Sets) {
		setRefs := []*firewallconfigs.HTTPFirewallRuleSetRef{}
		err = json.Unmarshal(group.Sets, &setRefs)
		if err != nil {
			return nil, err
		}
		for _, setRef := range setRefs {
			setConfig, err := SharedHTTPFirewallRuleSetDAO.ComposeFirewallRuleSet(tx, setRef.SetId)
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

// CreateGroupFromConfig 从配置中创建分组
func (this *HTTPFirewallRuleGroupDAO) CreateGroupFromConfig(tx *dbs.Tx, groupConfig *firewallconfigs.HTTPFirewallRuleGroup) (int64, error) {
	op := NewHTTPFirewallRuleGroupOperator()
	op.IsOn = groupConfig.IsOn
	op.Name = groupConfig.Name
	op.Description = groupConfig.Description
	op.State = HTTPFirewallRuleGroupStateEnabled
	op.Code = groupConfig.Code
	op.IsTemplate = groupConfig.IsTemplate

	// sets
	setRefs := []*firewallconfigs.HTTPFirewallRuleSetRef{}
	for _, setConfig := range groupConfig.Sets {
		setId, err := SharedHTTPFirewallRuleSetDAO.CreateOrUpdateSetFromConfig(tx, setConfig)
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
	err = this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateGroupIsOn 修改开启状态
func (this *HTTPFirewallRuleGroupDAO) UpdateGroupIsOn(tx *dbs.Tx, groupId int64, isOn bool) error {
	_, err := this.Query(tx).
		Pk(groupId).
		Set("isOn", isOn).
		Update()
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, groupId)
}

// CreateGroup 创建分组
func (this *HTTPFirewallRuleGroupDAO) CreateGroup(tx *dbs.Tx, isOn bool, name string, code string, description string) (int64, error) {
	op := NewHTTPFirewallRuleGroupOperator()
	op.State = HTTPFirewallRuleStateEnabled
	op.IsOn = isOn
	op.Name = name
	op.Code = code
	op.Description = description
	err := this.Save(tx, op)
	if err != nil {
		return 0, err
	}
	return types.Int64(op.Id), nil
}

// UpdateGroup 修改分组
func (this *HTTPFirewallRuleGroupDAO) UpdateGroup(tx *dbs.Tx, groupId int64, isOn bool, name string, code string, description string) error {
	if groupId <= 0 {
		return errors.New("invalid groupId")
	}
	op := NewHTTPFirewallRuleGroupOperator()
	op.Id = groupId
	op.IsOn = isOn
	op.Name = name
	op.Code = code
	op.Description = description
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, groupId)
}

// UpdateGroupSets 修改分组中的规则集
func (this *HTTPFirewallRuleGroupDAO) UpdateGroupSets(tx *dbs.Tx, groupId int64, setRefsJSON []byte) error {
	if groupId <= 0 {
		return errors.New("invalid groupId")
	}
	op := NewHTTPFirewallRuleGroupOperator()
	op.Id = groupId
	op.Sets = setRefsJSON
	err := this.Save(tx, op)
	if err != nil {
		return err
	}
	return this.NotifyUpdate(tx, groupId)
}

// FindRuleGroupIdWithRuleSetId 根据规则集查找规则分组
func (this *HTTPFirewallRuleGroupDAO) FindRuleGroupIdWithRuleSetId(tx *dbs.Tx, setId int64) (int64, error) {
	return this.Query(tx).
		State(HTTPFirewallRuleStateEnabled).
		Where("JSON_CONTAINS(sets, :jsonQuery)").
		Param("jsonQuery", maps.Map{"setId": setId}.AsJSON()).
		ResultPk().
		FindInt64Col(0)
}

// CheckUserRuleGroup 检查用户所属分组
func (this *HTTPFirewallRuleGroupDAO) CheckUserRuleGroup(tx *dbs.Tx, userId int64, groupId int64) error {
	policyId, err := SharedHTTPFirewallPolicyDAO.FindEnabledFirewallPolicyIdWithRuleGroupId(tx, groupId)
	if err != nil {
		return err
	}
	if policyId == 0 {
		return ErrNotFound
	}
	return SharedHTTPFirewallPolicyDAO.CheckUserFirewallPolicy(tx, userId, policyId)
}

// NotifyUpdate 通知更新
func (this *HTTPFirewallRuleGroupDAO) NotifyUpdate(tx *dbs.Tx, groupId int64) error {
	policyId, err := SharedHTTPFirewallPolicyDAO.FindEnabledFirewallPolicyIdWithRuleGroupId(tx, groupId)
	if err != nil {
		return err
	}
	if policyId > 0 {
		return SharedHTTPFirewallPolicyDAO.NotifyUpdate(tx, policyId)
	}
	return nil
}
