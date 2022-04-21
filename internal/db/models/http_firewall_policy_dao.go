package models

import (
	"encoding/json"
	dbutils "github.com/TeaOSLab/EdgeAPI/internal/db/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/maps"
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

var SharedHTTPFirewallPolicyDAO *HTTPFirewallPolicyDAO

func init() {
	dbs.OnReady(func() {
		SharedHTTPFirewallPolicyDAO = NewHTTPFirewallPolicyDAO()
	})
}

// Init 初始化
func (this *HTTPFirewallPolicyDAO) Init() {
	_ = this.DAOObject.Init()
}

// EnableHTTPFirewallPolicy 启用条目
func (this *HTTPFirewallPolicyDAO) EnableHTTPFirewallPolicy(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", HTTPFirewallPolicyStateEnabled).
		Update()
	return err
}

// DisableHTTPFirewallPolicy 禁用条目
func (this *HTTPFirewallPolicyDAO) DisableHTTPFirewallPolicy(tx *dbs.Tx, policyId int64) error {
	_, err := this.Query(tx).
		Pk(policyId).
		Set("state", HTTPFirewallPolicyStateDisabled).
		Update()
	if err != nil {
		return err
	}

	err = this.NotifyDisable(tx, policyId)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, policyId)
}

// FindEnabledHTTPFirewallPolicy 查找启用中的条目
func (this *HTTPFirewallPolicyDAO) FindEnabledHTTPFirewallPolicy(tx *dbs.Tx, id int64) (*HTTPFirewallPolicy, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", HTTPFirewallPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallPolicy), err
}

// FindHTTPFirewallPolicyName 根据主键查找名称
func (this *HTTPFirewallPolicyDAO) FindHTTPFirewallPolicyName(tx *dbs.Tx, id int64) (string, error) {
	return this.Query(tx).
		Pk(id).
		Result("name").
		FindStringCol("")
}

// FindEnabledHTTPFirewallPolicyBasic 获取WAF策略基本信息
func (this *HTTPFirewallPolicyDAO) FindEnabledHTTPFirewallPolicyBasic(tx *dbs.Tx, policyId int64) (*HTTPFirewallPolicy, error) {
	result, err := this.Query(tx).
		Pk(policyId).
		Result("id", "name", "serverId", "isOn").
		Attr("state", HTTPFirewallPolicyStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*HTTPFirewallPolicy), err
}

// FindAllEnabledFirewallPolicies 查找所有可用策略
func (this *HTTPFirewallPolicyDAO) FindAllEnabledFirewallPolicies(tx *dbs.Tx) (result []*HTTPFirewallPolicy, err error) {
	_, err = this.Query(tx).
		State(HTTPFirewallPolicyStateEnabled).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// CreateFirewallPolicy 创建策略
func (this *HTTPFirewallPolicyDAO) CreateFirewallPolicy(tx *dbs.Tx, userId int64, serverGroupId int64, serverId int64, isOn bool, name string, description string, inboundJSON []byte, outboundJSON []byte) (int64, error) {
	op := NewHTTPFirewallPolicyOperator()
	op.UserId = userId
	op.GroupId = serverGroupId
	op.ServerId = serverId
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
	op.UseLocalFirewall = true

	{
		synFloodJSON, err := json.Marshal(firewallconfigs.DefaultSYNFloodConfig())
		if err != nil {
			return 0, err
		}
		op.SynFlood = synFloodJSON
	}

	err := this.Save(tx, op)
	return types.Int64(op.Id), err
}

// CreateDefaultFirewallPolicy 创建默认的WAF策略
func (this *HTTPFirewallPolicyDAO) CreateDefaultFirewallPolicy(tx *dbs.Tx, name string) (int64, error) {
	policyId, err := this.CreateFirewallPolicy(tx, 0, 0, 0, true, "\""+name+"\"WAF策略", "默认创建的WAF策略", nil, nil)
	if err != nil {
		return 0, err
	}

	// 初始化
	var groupCodes = []string{}

	templatePolicy := firewallconfigs.HTTPFirewallTemplate()
	for _, group := range templatePolicy.AllRuleGroups() {
		groupCodes = append(groupCodes, group.Code)
	}

	inboundConfig := &firewallconfigs.HTTPFirewallInboundConfig{IsOn: true}
	outboundConfig := &firewallconfigs.HTTPFirewallOutboundConfig{IsOn: true}
	if templatePolicy.Inbound != nil {
		for _, group := range templatePolicy.Inbound.Groups {
			isOn := lists.ContainsString(groupCodes, group.Code)
			group.IsOn = isOn

			groupId, err := SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(tx, group)
			if err != nil {
				return 0, err
			}
			inboundConfig.GroupRefs = append(inboundConfig.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
				IsOn:    true,
				GroupId: groupId,
			})
		}
	}
	if templatePolicy.Outbound != nil {
		for _, group := range templatePolicy.Outbound.Groups {
			isOn := lists.ContainsString(groupCodes, group.Code)
			group.IsOn = isOn

			groupId, err := SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(tx, group)
			if err != nil {
				return 0, err
			}
			outboundConfig.GroupRefs = append(outboundConfig.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
				IsOn:    true,
				GroupId: groupId,
			})
		}
	}

	inboundConfigJSON, err := json.Marshal(inboundConfig)
	if err != nil {
		return 0, err
	}

	outboundConfigJSON, err := json.Marshal(outboundConfig)
	if err != nil {
		return 0, err
	}

	err = this.UpdateFirewallPolicyInboundAndOutbound(tx, policyId, inboundConfigJSON, outboundConfigJSON, false)
	if err != nil {
		return 0, err
	}
	return policyId, nil
}

// UpdateFirewallPolicyInboundAndOutbound 修改策略的Inbound和Outbound
func (this *HTTPFirewallPolicyDAO) UpdateFirewallPolicyInboundAndOutbound(tx *dbs.Tx, policyId int64, inboundJSON []byte, outboundJSON []byte, shouldNotify bool) error {
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
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	if shouldNotify {
		return this.NotifyUpdate(tx, policyId)
	}

	return nil
}

// UpdateFirewallPolicyInbound 修改策略的Inbound
func (this *HTTPFirewallPolicyDAO) UpdateFirewallPolicyInbound(tx *dbs.Tx, policyId int64, inboundJSON []byte) error {
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
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, policyId)
}

// UpdateFirewallPolicy 修改策略
func (this *HTTPFirewallPolicyDAO) UpdateFirewallPolicy(tx *dbs.Tx, policyId int64, isOn bool, name string, description string, inboundJSON []byte, outboundJSON []byte, blockOptionsJSON []byte, mode firewallconfigs.FirewallMode, useLocalFirewall bool, synFloodConfig *firewallconfigs.SYNFloodConfig) error {
	if policyId <= 0 {
		return errors.New("invalid policyId")
	}
	op := NewHTTPFirewallPolicyOperator()
	op.Id = policyId
	op.IsOn = isOn
	op.Name = name
	op.Description = description
	op.Mode = mode
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
	if len(blockOptionsJSON) > 0 {
		op.BlockOptions = blockOptionsJSON
	}

	if synFloodConfig != nil {
		synFloodConfigJSON, err := json.Marshal(synFloodConfig)
		if err != nil {
			return err
		}
		op.SynFlood = synFloodConfigJSON
	} else {
		op.SynFlood = "null"
	}

	op.UseLocalFirewall = useLocalFirewall
	err := this.Save(tx, op)
	if err != nil {
		return err
	}

	return this.NotifyUpdate(tx, policyId)
}

// CountAllEnabledFirewallPolicies 计算所有可用的策略数量
func (this *HTTPFirewallPolicyDAO) CountAllEnabledFirewallPolicies(tx *dbs.Tx, clusterId int64, keyword string) (int64, error) {
	query := this.Query(tx)
	if clusterId > 0 {
		query.Where("id IN (SELECT httpFirewallPolicyId FROM " + SharedNodeClusterDAO.Table + " WHERE id=:clusterId)")
		query.Param("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	return query.
		State(HTTPFirewallPolicyStateEnabled).
		Attr("userId", 0).
		Attr("serverId", 0).
		Attr("groupId", 0).
		Count()
}

// ListEnabledFirewallPolicies 列出单页的策略
func (this *HTTPFirewallPolicyDAO) ListEnabledFirewallPolicies(tx *dbs.Tx, clusterId int64, keyword string, offset int64, size int64) (result []*HTTPFirewallPolicy, err error) {
	query := this.Query(tx)
	if clusterId > 0 {
		query.Where("id IN (SELECT httpFirewallPolicyId FROM " + SharedNodeClusterDAO.Table + " WHERE id=:clusterId)")
		query.Param("clusterId", clusterId)
	}
	if len(keyword) > 0 {
		query.Where("(name LIKE :keyword)").
			Param("keyword", dbutils.QuoteLike(keyword))
	}
	_, err = query.
		State(HTTPFirewallPolicyStateEnabled).
		Attr("userId", 0).
		Attr("serverId", 0).
		Attr("groupId", 0).
		Offset(offset).
		Limit(size).
		DescPk().
		Slice(&result).
		FindAll()
	return
}

// ComposeFirewallPolicy 组合策略配置
func (this *HTTPFirewallPolicyDAO) ComposeFirewallPolicy(tx *dbs.Tx, policyId int64, cacheMap *utils.CacheMap) (*firewallconfigs.HTTPFirewallPolicy, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	var cacheKey = this.Table + ":config:" + types.String(policyId)
	var cache, _ = cacheMap.Get(cacheKey)
	if cache != nil {
		return cache.(*firewallconfigs.HTTPFirewallPolicy), nil
	}

	policy, err := this.FindEnabledHTTPFirewallPolicy(tx, policyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}

	var config = &firewallconfigs.HTTPFirewallPolicy{}
	config.Id = int64(policy.Id)
	config.IsOn = policy.IsOn
	config.Name = policy.Name
	config.Description = policy.Description
	config.UseLocalFirewall = policy.UseLocalFirewall == 1

	if len(policy.Mode) == 0 {
		policy.Mode = firewallconfigs.FirewallModeDefend
	}
	config.Mode = policy.Mode

	// Inbound
	inbound := &firewallconfigs.HTTPFirewallInboundConfig{}
	if IsNotNull(policy.Inbound) {
		err = json.Unmarshal(policy.Inbound, inbound)
		if err != nil {
			return nil, err
		}
		if len(inbound.GroupRefs) > 0 {
			resultGroupRefs := []*firewallconfigs.HTTPFirewallRuleGroupRef{}
			resultGroups := []*firewallconfigs.HTTPFirewallRuleGroup{}

			for _, groupRef := range inbound.GroupRefs {
				groupConfig, err := SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(tx, groupRef.GroupId)
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
	}
	config.Inbound = inbound

	// Outbound
	outbound := &firewallconfigs.HTTPFirewallOutboundConfig{}
	if IsNotNull(policy.Outbound) {
		err = json.Unmarshal(policy.Outbound, outbound)
		if err != nil {
			return nil, err
		}
		if len(outbound.GroupRefs) > 0 {
			resultGroupRefs := []*firewallconfigs.HTTPFirewallRuleGroupRef{}
			resultGroups := []*firewallconfigs.HTTPFirewallRuleGroup{}

			for _, groupRef := range outbound.GroupRefs {
				groupConfig, err := SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(tx, groupRef.GroupId)
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
	}
	config.Outbound = outbound

	// Block动作配置
	if IsNotNull(policy.BlockOptions) {
		blockAction := &firewallconfigs.HTTPFirewallBlockAction{}
		err = json.Unmarshal(policy.BlockOptions, blockAction)
		if err != nil {
			return config, err
		}
		config.BlockOptions = blockAction
	}

	// syn flood
	if IsNotNull(policy.SynFlood) {
		var synFloodConfig = &firewallconfigs.SYNFloodConfig{}
		err = json.Unmarshal(policy.SynFlood, synFloodConfig)
		if err != nil {
			return nil, err
		}
		config.SYNFlood = synFloodConfig
	}

	// log
	if IsNotNull(policy.Log) {
		var logConfig = &firewallconfigs.HTTPFirewallPolicyLogConfig{}
		err = json.Unmarshal(policy.Log, logConfig)
		if err != nil {
			return nil, err
		}
		config.Log = logConfig
	} else {
		config.Log = firewallconfigs.DefaultHTTPFirewallPolicyLogConfig
	}

	if cacheMap != nil {
		cacheMap.Put(cacheKey, config)
	}

	return config, nil
}

// CheckUserFirewallPolicy 检查用户防火墙策略
func (this *HTTPFirewallPolicyDAO) CheckUserFirewallPolicy(tx *dbs.Tx, userId int64, firewallPolicyId int64) error {
	ok, err := this.Query(tx).
		Pk(firewallPolicyId).
		Attr("userId", userId).
		Exist()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	// 检查是否为用户Server所使用
	webIds, err := SharedHTTPWebDAO.FindAllWebIdsWithHTTPFirewallPolicyId(tx, firewallPolicyId)
	if err != nil {
		return err
	}
	for _, webId := range webIds {
		err := SharedHTTPWebDAO.CheckUserWeb(tx, userId, webId)
		if err != nil {
			if err != ErrNotFound {
				return err
			}
		} else {
			return nil
		}
	}

	return ErrNotFound
}

// FindEnabledFirewallPolicyIdsWithIPListId 查找包含某个IPList的所有策略
func (this *HTTPFirewallPolicyDAO) FindEnabledFirewallPolicyIdsWithIPListId(tx *dbs.Tx, ipListId int64) ([]int64, error) {
	ones, err := this.Query(tx).
		ResultPk().
		State(HTTPFirewallPolicyStateEnabled).
		Where("(JSON_CONTAINS(inbound, :listQuery, '$.whiteListRef') OR JSON_CONTAINS(inbound, :listQuery, '$.blackListRef') OR JSON_CONTAINS(inbound, :listQuery, '$.publicWhiteListRefs')  OR JSON_CONTAINS(inbound, :listQuery, '$.publicBlackListRefs'))").
		Param("listQuery", maps.Map{"isOn": true, "listId": ipListId}.AsJSON()).
		FindAll()
	if err != nil {
		return nil, err
	}
	result := []int64{}
	for _, one := range ones {
		result = append(result, int64(one.(*HTTPFirewallPolicy).Id))
	}
	return result, nil
}

// FindEnabledFirewallPolicyWithIPListId 查找使用某个IPList的策略
func (this *HTTPFirewallPolicyDAO) FindEnabledFirewallPolicyWithIPListId(tx *dbs.Tx, ipListId int64) (*HTTPFirewallPolicy, error) {
	one, err := this.Query(tx).
		State(HTTPFirewallPolicyStateEnabled).
		Where("(JSON_CONTAINS(inbound, :listQuery, '$.whiteListRef') OR JSON_CONTAINS(inbound, :listQuery, '$.blackListRef'))").
		Param("listQuery", maps.Map{"isOn": true, "listId": ipListId}.AsJSON()).
		Find()
	if err != nil || one == nil {
		return nil, err
	}
	return one.(*HTTPFirewallPolicy), err
}

// FindEnabledFirewallPolicyIdWithRuleGroupId 查找包含某个规则分组的策略ID
func (this *HTTPFirewallPolicyDAO) FindEnabledFirewallPolicyIdWithRuleGroupId(tx *dbs.Tx, ruleGroupId int64) (int64, error) {
	return this.Query(tx).
		ResultPk().
		State(HTTPFirewallPolicyStateEnabled).
		Where("(JSON_CONTAINS(inbound, :jsonQuery, '$.groupRefs') OR JSON_CONTAINS(outbound, :jsonQuery, '$.groupRefs'))").
		Param("jsonQuery", maps.Map{"groupId": ruleGroupId}.AsJSON()).
		FindInt64Col(0)
}

// UpdateFirewallPolicyServerId 设置某个策略所属的服务ID
func (this *HTTPFirewallPolicyDAO) UpdateFirewallPolicyServerId(tx *dbs.Tx, policyId int64, serverId int64) error {
	_, err := this.Query(tx).
		Pk(policyId).
		Set("serverId", serverId).
		Update()
	return err
}

// FindFirewallPolicyIdsWithServerId 查找服务独立关联的策略IDs
func (this *HTTPFirewallPolicyDAO) FindFirewallPolicyIdsWithServerId(tx *dbs.Tx, serverId int64) ([]int64, error) {
	var result = []int64{}
	ones, err := this.Query(tx).
		Attr("serverId", serverId).
		State(HTTPFirewallPolicyStateEnabled).
		Result("id").
		FindAll()
	if err != nil {
		return nil, err
	}
	for _, one := range ones {
		result = append(result, int64(one.(*HTTPFirewallPolicy).Id))
	}
	return result, nil
}

// NotifyUpdate 通知更新
func (this *HTTPFirewallPolicyDAO) NotifyUpdate(tx *dbs.Tx, policyId int64) error {
	webIds, err := SharedHTTPWebDAO.FindAllWebIdsWithHTTPFirewallPolicyId(tx, policyId)
	if err != nil {
		return err
	}
	for _, webId := range webIds {
		err := SharedHTTPWebDAO.NotifyUpdate(tx, webId)
		if err != nil {
			return err
		}
	}

	clusterIds, err := SharedNodeClusterDAO.FindAllEnabledNodeClusterIdsWithHTTPFirewallPolicyId(tx, policyId)
	if err != nil {
		return err
	}
	for _, clusterId := range clusterIds {
		err := SharedNodeClusterDAO.NotifyUpdate(tx, clusterId)
		if err != nil {
			return err
		}
	}

	return nil
}

// NotifyDisable 通知禁用
func (this *HTTPFirewallPolicyDAO) NotifyDisable(tx *dbs.Tx, policyId int64) error {
	if policyId <= 0 {
		return nil
	}

	// 禁用IP名单
	inboundString, err := this.Query(tx).
		Pk(policyId).
		Result("inbound").
		FindStringCol("")
	if err != nil {
		return err
	}
	if len(inboundString) > 0 {
		var inboundConfig = &firewallconfigs.HTTPFirewallInboundConfig{}
		err = json.Unmarshal([]byte(inboundString), inboundConfig)
		if err != nil {
			// 不处理错误
			return nil
		}

		if inboundConfig.AllowListRef != nil && inboundConfig.AllowListRef.ListId > 0 {
			err = SharedIPListDAO.DisableIPList(tx, inboundConfig.AllowListRef.ListId)
			if err != nil {
				return err
			}

			err = SharedIPItemDAO.DisableIPItemsWithListId(tx, inboundConfig.AllowListRef.ListId)
			if err != nil {
				return err
			}
		}

		if inboundConfig.DenyListRef != nil && inboundConfig.DenyListRef.ListId > 0 {
			err = SharedIPListDAO.DisableIPList(tx, inboundConfig.DenyListRef.ListId)
			if err != nil {
				return err
			}

			err = SharedIPItemDAO.DisableIPItemsWithListId(tx, inboundConfig.DenyListRef.ListId)
			if err != nil {
				return err
			}
		}

		if inboundConfig.GreyListRef != nil && inboundConfig.GreyListRef.ListId > 0 {
			err = SharedIPListDAO.DisableIPList(tx, inboundConfig.GreyListRef.ListId)
			if err != nil {
				return err
			}

			err = SharedIPItemDAO.DisableIPItemsWithListId(tx, inboundConfig.GreyListRef.ListId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
