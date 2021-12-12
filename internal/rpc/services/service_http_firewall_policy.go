package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/iplibrary"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/iwind/TeaGo/lists"
	"net"
)

// HTTPFirewallPolicyService HTTP防火墙（WAF）相关服务
type HTTPFirewallPolicyService struct {
	BaseService
}

// FindAllEnabledHTTPFirewallPolicies 获取所有可用策略
func (this *HTTPFirewallPolicyService) FindAllEnabledHTTPFirewallPolicies(ctx context.Context, req *pb.FindAllEnabledHTTPFirewallPoliciesRequest) (*pb.FindAllEnabledHTTPFirewallPoliciesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	policies, err := models.SharedHTTPFirewallPolicyDAO.FindAllEnabledFirewallPolicies(tx)
	if err != nil {
		return nil, err
	}

	result := []*pb.HTTPFirewallPolicy{}
	for _, p := range policies {
		result = append(result, &pb.HTTPFirewallPolicy{
			Id:           int64(p.Id),
			Name:         p.Name,
			Description:  p.Description,
			IsOn:         p.IsOn == 1,
			InboundJSON:  []byte(p.Inbound),
			OutboundJSON: []byte(p.Outbound),
			Mode:         p.Mode,
		})
	}

	return &pb.FindAllEnabledHTTPFirewallPoliciesResponse{FirewallPolicies: result}, nil
}

// CreateHTTPFirewallPolicy 创建防火墙策略
func (this *HTTPFirewallPolicyService) CreateHTTPFirewallPolicy(ctx context.Context, req *pb.CreateHTTPFirewallPolicyRequest) (*pb.CreateHTTPFirewallPolicyResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	policyId, err := models.SharedHTTPFirewallPolicyDAO.CreateFirewallPolicy(tx, userId, req.ServerGroupId, req.ServerId, req.IsOn, req.Name, req.Description, nil, nil)
	if err != nil {
		return nil, err
	}

	// 初始化
	inboundConfig := &firewallconfigs.HTTPFirewallInboundConfig{IsOn: true}
	outboundConfig := &firewallconfigs.HTTPFirewallOutboundConfig{IsOn: true}
	templatePolicy := firewallconfigs.HTTPFirewallTemplate()
	if templatePolicy.Inbound != nil {
		for _, group := range templatePolicy.Inbound.Groups {
			isOn := lists.ContainsString(req.HttpFirewallGroupCodes, group.Code)
			group.IsOn = isOn

			groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(tx, group)
			if err != nil {
				return nil, err
			}
			inboundConfig.GroupRefs = append(inboundConfig.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
				IsOn:    true,
				GroupId: groupId,
			})
		}
	}
	if templatePolicy.Outbound != nil {
		for _, group := range templatePolicy.Outbound.Groups {
			isOn := lists.ContainsString(req.HttpFirewallGroupCodes, group.Code)
			group.IsOn = isOn

			groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(tx, group)
			if err != nil {
				return nil, err
			}
			outboundConfig.GroupRefs = append(outboundConfig.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
				IsOn:    true,
				GroupId: groupId,
			})
		}
	}

	inboundConfigJSON, err := json.Marshal(inboundConfig)
	if err != nil {
		return nil, err
	}

	outboundConfigJSON, err := json.Marshal(outboundConfig)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(tx, policyId, inboundConfigJSON, outboundConfigJSON, false)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPFirewallPolicyResponse{HttpFirewallPolicyId: policyId}, nil
}

// CreateEmptyHTTPFirewallPolicy 创建空防火墙策略
func (this *HTTPFirewallPolicyService) CreateEmptyHTTPFirewallPolicy(ctx context.Context, req *pb.CreateEmptyHTTPFirewallPolicyRequest) (*pb.CreateEmptyHTTPFirewallPolicyResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}
	}

	tx := this.NullTx()

	policyId, err := models.SharedHTTPFirewallPolicyDAO.CreateFirewallPolicy(tx, userId, req.ServerGroupId, req.ServerId, req.IsOn, req.Name, req.Description, nil, nil)
	if err != nil {
		return nil, err
	}

	// 初始化
	inboundConfig := &firewallconfigs.HTTPFirewallInboundConfig{IsOn: true}
	outboundConfig := &firewallconfigs.HTTPFirewallOutboundConfig{IsOn: true}

	inboundConfigJSON, err := json.Marshal(inboundConfig)
	if err != nil {
		return nil, err
	}

	outboundConfigJSON, err := json.Marshal(outboundConfig)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(tx, policyId, inboundConfigJSON, outboundConfigJSON, false)
	if err != nil {
		return nil, err
	}

	return &pb.CreateEmptyHTTPFirewallPolicyResponse{HttpFirewallPolicyId: policyId}, nil
}

// UpdateHTTPFirewallPolicy 修改防火墙策略
func (this *HTTPFirewallPolicyService) UpdateHTTPFirewallPolicy(ctx context.Context, req *pb.UpdateHTTPFirewallPolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	templatePolicy := firewallconfigs.HTTPFirewallTemplate()

	tx := this.NullTx()

	// 已经有的数据
	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, req.HttpFirewallPolicyId, nil)
	if err != nil {
		return nil, err
	}
	if firewallPolicy == nil {
		return nil, errors.New("can not found firewall policy")
	}

	inboundConfig := firewallPolicy.Inbound
	if inboundConfig == nil {
		inboundConfig = &firewallconfigs.HTTPFirewallInboundConfig{IsOn: true}
	}

	outboundConfig := firewallPolicy.Outbound
	if outboundConfig == nil {
		outboundConfig = &firewallconfigs.HTTPFirewallOutboundConfig{IsOn: true}
	}

	// 更新老的
	oldCodes := []string{}
	if firewallPolicy.Inbound != nil {
		for _, g := range firewallPolicy.Inbound.Groups {
			if len(g.Code) > 0 {
				oldCodes = append(oldCodes, g.Code)
				if lists.ContainsString(req.FirewallGroupCodes, g.Code) {
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(tx, g.Id, true)
					if err != nil {
						return nil, err
					}
				} else {
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(tx, g.Id, false)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	if firewallPolicy.Outbound != nil {
		for _, g := range firewallPolicy.Outbound.Groups {
			if len(g.Code) > 0 {
				oldCodes = append(oldCodes, g.Code)
				if lists.ContainsString(req.FirewallGroupCodes, g.Code) {
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(tx, g.Id, true)
					if err != nil {
						return nil, err
					}
				} else {
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(tx, g.Id, false)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	// 加入新的
	if templatePolicy.Inbound != nil {
		for _, group := range templatePolicy.Inbound.Groups {
			if lists.ContainsString(oldCodes, group.Code) {
				continue
			}

			isOn := lists.ContainsString(req.FirewallGroupCodes, group.Code)
			group.IsOn = isOn

			groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(tx, group)
			if err != nil {
				return nil, err
			}
			inboundConfig.GroupRefs = append(inboundConfig.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
				IsOn:    true,
				GroupId: groupId,
			})
		}
	}
	if templatePolicy.Outbound != nil {
		for _, group := range templatePolicy.Outbound.Groups {
			if lists.ContainsString(oldCodes, group.Code) {
				continue
			}

			isOn := lists.ContainsString(req.FirewallGroupCodes, group.Code)
			group.IsOn = isOn

			groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(tx, group)
			if err != nil {
				return nil, err
			}
			outboundConfig.GroupRefs = append(outboundConfig.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
				IsOn:    true,
				GroupId: groupId,
			})
		}
	}

	inboundConfigJSON, err := json.Marshal(inboundConfig)
	if err != nil {
		return nil, err
	}

	outboundConfigJSON, err := json.Marshal(outboundConfig)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicy(tx, req.HttpFirewallPolicyId, req.IsOn, req.Name, req.Description, inboundConfigJSON, outboundConfigJSON, req.BlockOptionsJSON, req.Mode)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPFirewallPolicyGroups 修改分组信息
func (this *HTTPFirewallPolicyService) UpdateHTTPFirewallPolicyGroups(ctx context.Context, req *pb.UpdateHTTPFirewallPolicyGroupsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPFirewallPolicyDAO.CheckUserFirewallPolicy(nil, userId, req.HttpFirewallPolicyId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(tx, req.HttpFirewallPolicyId, req.InboundJSON, req.OutboundJSON, true)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPFirewallInboundConfig 修改inbound信息
func (this *HTTPFirewallPolicyService) UpdateHTTPFirewallInboundConfig(ctx context.Context, req *pb.UpdateHTTPFirewallInboundConfigRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	if userId > 0 {
		err = models.SharedHTTPFirewallPolicyDAO.CheckUserFirewallPolicy(tx, userId, req.HttpFirewallPolicyId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInbound(tx, req.HttpFirewallPolicyId, req.InboundJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountAllEnabledHTTPFirewallPolicies 计算可用的防火墙策略数量
func (this *HTTPFirewallPolicyService) CountAllEnabledHTTPFirewallPolicies(ctx context.Context, req *pb.CountAllEnabledHTTPFirewallPoliciesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedHTTPFirewallPolicyDAO.CountAllEnabledFirewallPolicies(tx, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledHTTPFirewallPolicies 列出单页的防火墙策略
func (this *HTTPFirewallPolicyService) ListEnabledHTTPFirewallPolicies(ctx context.Context, req *pb.ListEnabledHTTPFirewallPoliciesRequest) (*pb.ListEnabledHTTPFirewallPoliciesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	policies, err := models.SharedHTTPFirewallPolicyDAO.ListEnabledFirewallPolicies(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.HTTPFirewallPolicy{}
	for _, p := range policies {
		result = append(result, &pb.HTTPFirewallPolicy{
			Id:           int64(p.Id),
			Name:         p.Name,
			Description:  p.Description,
			IsOn:         p.IsOn == 1,
			InboundJSON:  []byte(p.Inbound),
			OutboundJSON: []byte(p.Outbound),
			Mode:         p.Mode,
		})
	}

	return &pb.ListEnabledHTTPFirewallPoliciesResponse{HttpFirewallPolicies: result}, nil
}

// DeleteHTTPFirewallPolicy 删除某个防火墙策略
func (this *HTTPFirewallPolicyService) DeleteHTTPFirewallPolicy(ctx context.Context, req *pb.DeleteHTTPFirewallPolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedHTTPFirewallPolicyDAO.DisableHTTPFirewallPolicy(tx, req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledHTTPFirewallPolicyConfig 查找单个防火墙配置
func (this *HTTPFirewallPolicyService) FindEnabledHTTPFirewallPolicyConfig(ctx context.Context, req *pb.FindEnabledHTTPFirewallPolicyConfigRequest) (*pb.FindEnabledHTTPFirewallPolicyConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 校验权限
		err = models.SharedHTTPFirewallPolicyDAO.CheckUserFirewallPolicy(nil, userId, req.HttpFirewallPolicyId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	config, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, req.HttpFirewallPolicyId, nil)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return &pb.FindEnabledHTTPFirewallPolicyConfigResponse{HttpFirewallPolicyJSON: nil}, nil
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledHTTPFirewallPolicyConfigResponse{HttpFirewallPolicyJSON: configJSON}, nil
}

// FindEnabledHTTPFirewallPolicy 获取防火墙的基本信息
func (this *HTTPFirewallPolicyService) FindEnabledHTTPFirewallPolicy(ctx context.Context, req *pb.FindEnabledHTTPFirewallPolicyRequest) (*pb.FindEnabledHTTPFirewallPolicyResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPFirewallPolicyDAO.CheckUserFirewallPolicy(nil, userId, req.HttpFirewallPolicyId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	policy, err := models.SharedHTTPFirewallPolicyDAO.FindEnabledHTTPFirewallPolicy(tx, req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return &pb.FindEnabledHTTPFirewallPolicyResponse{HttpFirewallPolicy: nil}, nil
	}
	return &pb.FindEnabledHTTPFirewallPolicyResponse{HttpFirewallPolicy: &pb.HTTPFirewallPolicy{
		Id:           int64(policy.Id),
		Name:         policy.Name,
		Description:  policy.Description,
		IsOn:         policy.IsOn == 1,
		InboundJSON:  []byte(policy.Inbound),
		OutboundJSON: []byte(policy.Outbound),
		Mode:         policy.Mode,
	}}, nil
}

// ImportHTTPFirewallPolicy 导入策略数据
func (this *HTTPFirewallPolicyService) ImportHTTPFirewallPolicy(ctx context.Context, req *pb.ImportHTTPFirewallPolicyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	tx := this.NullTx()

	oldConfig, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, req.HttpFirewallPolicyId, nil)
	if err != nil {
		return nil, err
	}
	if oldConfig == nil {
		return nil, errors.New("can not find policy")
	}

	// 解析数据
	newConfig := &firewallconfigs.HTTPFirewallPolicy{}
	err = json.Unmarshal(req.HttpFirewallPolicyJSON, newConfig)
	if err != nil {
		return nil, err
	}

	// 入站分组
	if newConfig.Inbound != nil {
		for _, g := range newConfig.Inbound.Groups {
			var oldGroup *firewallconfigs.HTTPFirewallRuleGroup

			// 使用代号查找
			if len(g.Code) > 0 {
				oldGroup = oldConfig.FindRuleGroupWithCode(g.Code)
			}

			// 再次根据Name查找
			if oldGroup == nil && len(g.Name) > 0 {
				oldGroup = oldConfig.FindRuleGroupWithName(g.Name)
			}

			if oldGroup == nil {
				// 新创建分组
				groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(tx, g)
				if err != nil {
					return nil, err
				}
				oldConfig.Inbound.GroupRefs = append(oldConfig.Inbound.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
					IsOn:    true,
					GroupId: groupId,
				})
			} else {
				setRefs := []*firewallconfigs.HTTPFirewallRuleSetRef{}
				for _, set := range g.Sets {
					setId, err := models.SharedHTTPFirewallRuleSetDAO.CreateOrUpdateSetFromConfig(tx, set)
					if err != nil {
						return nil, err
					}
					setRefs = append(setRefs, &firewallconfigs.HTTPFirewallRuleSetRef{
						IsOn:  true,
						SetId: setId,
					})
				}
				setsJSON, err := json.Marshal(setRefs)
				if err != nil {
					return nil, err
				}

				err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroup(tx, oldGroup.Id, g.IsOn, g.Name, g.Code, g.Description)
				if err != nil {
					return nil, err
				}

				err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupSets(tx, oldGroup.Id, setsJSON)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// 出站分组
	if newConfig.Outbound != nil {
		for _, g := range newConfig.Outbound.Groups {
			var oldGroup *firewallconfigs.HTTPFirewallRuleGroup

			// 使用代号查找
			if len(g.Code) > 0 {
				oldGroup = oldConfig.FindRuleGroupWithCode(g.Code)
			}

			// 再次根据Name查找
			if oldGroup == nil && len(g.Name) > 0 {
				oldGroup = oldConfig.FindRuleGroupWithName(g.Name)
			}

			if oldGroup == nil {
				// 新创建分组
				groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(tx, g)
				if err != nil {
					return nil, err
				}
				oldConfig.Outbound.GroupRefs = append(oldConfig.Outbound.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
					IsOn:    true,
					GroupId: groupId,
				})
			} else {
				setRefs := []*firewallconfigs.HTTPFirewallRuleSetRef{}
				for _, set := range g.Sets {
					setId, err := models.SharedHTTPFirewallRuleSetDAO.CreateOrUpdateSetFromConfig(tx, set)
					if err != nil {
						return nil, err
					}
					setRefs = append(setRefs, &firewallconfigs.HTTPFirewallRuleSetRef{
						IsOn:  true,
						SetId: setId,
					})
				}
				setsJSON, err := json.Marshal(setRefs)
				if err != nil {
					return nil, err
				}
				err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroup(tx, oldGroup.Id, g.IsOn, g.Name, g.Code, g.Description)
				if err != nil {
					return nil, err
				}
				err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupSets(tx, oldGroup.Id, setsJSON)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// 保存Inbound和Outbound
	oldConfig.Inbound.Groups = nil
	oldConfig.Outbound.Groups = nil

	inboundJSON, err := json.Marshal(oldConfig.Inbound)
	if err != nil {
		return nil, err
	}

	outboundJSON, err := json.Marshal(oldConfig.Outbound)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(tx, req.HttpFirewallPolicyId, inboundJSON, outboundJSON, true)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CheckHTTPFirewallPolicyIPStatus 检查IP状态
func (this *HTTPFirewallPolicyService) CheckHTTPFirewallPolicyIPStatus(ctx context.Context, req *pb.CheckHTTPFirewallPolicyIPStatusRequest) (*pb.CheckHTTPFirewallPolicyIPStatusResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// 校验IP
	ip := net.ParseIP(req.Ip)
	if len(ip) == 0 {
		return &pb.CheckHTTPFirewallPolicyIPStatusResponse{
			IsOk:  false,
			Error: "请输入正确的IP",
		}, nil
	}
	ipLong := utils.IP2Long(req.Ip)

	tx := this.NullTx()
	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(tx, req.HttpFirewallPolicyId, nil)
	if err != nil {
		return nil, err
	}
	if firewallPolicy == nil {
		return &pb.CheckHTTPFirewallPolicyIPStatusResponse{
			IsOk:  false,
			Error: "找不到策略信息",
		}, nil
	}

	// 检查白名单
	if firewallPolicy.Inbound != nil &&
		firewallPolicy.Inbound.IsOn &&
		firewallPolicy.Inbound.AllowListRef != nil &&
		firewallPolicy.Inbound.AllowListRef.IsOn &&
		firewallPolicy.Inbound.AllowListRef.ListId > 0 {

		var listIds = []int64{}
		if firewallPolicy.Inbound.AllowListRef.ListId > 0 {
			listIds = append(listIds, firewallPolicy.Inbound.AllowListRef.ListId)
		}
		if len(firewallPolicy.Inbound.PublicAllowListRefs) > 0 {
			for _, ref := range firewallPolicy.Inbound.PublicAllowListRefs {
				if !ref.IsOn {
					continue
				}

				listIds = append(listIds, ref.ListId)
			}
		}

		for _, listId := range listIds {
			item, err := models.SharedIPItemDAO.FindEnabledItemContainsIP(tx, listId, ipLong)
			if err != nil {
				return nil, err
			}
			if item != nil {
				listName, err := models.SharedIPListDAO.FindIPListName(tx, listId)
				if err != nil {
					return nil, err
				}
				if len(listName) == 0 {
					listName = "白名单"
				}
				return &pb.CheckHTTPFirewallPolicyIPStatusResponse{
					IsOk:      true,
					Error:     "",
					IsFound:   true,
					IsAllowed: true,
					IpList:    &pb.IPList{Name: listName, Id: listId},
					IpItem: &pb.IPItem{
						Id:         int64(item.Id),
						IpFrom:     item.IpFrom,
						IpTo:       item.IpTo,
						ExpiredAt:  int64(item.ExpiredAt),
						Reason:     item.Reason,
						Type:       item.Type,
						EventLevel: item.EventLevel,
					},
					RegionCountry:  nil,
					RegionProvince: nil,
				}, nil
			}
		}
	}

	// 检查黑名单
	if firewallPolicy.Inbound != nil &&
		firewallPolicy.Inbound.IsOn &&
		firewallPolicy.Inbound.DenyListRef != nil &&
		firewallPolicy.Inbound.DenyListRef.IsOn &&
		firewallPolicy.Inbound.DenyListRef.ListId > 0 {
		var listIds = []int64{}
		if firewallPolicy.Inbound.DenyListRef.ListId > 0 {
			listIds = append(listIds, firewallPolicy.Inbound.DenyListRef.ListId)
		}
		if len(firewallPolicy.Inbound.PublicDenyListRefs) > 0 {
			for _, ref := range firewallPolicy.Inbound.PublicDenyListRefs {
				if !ref.IsOn {
					continue
				}

				listIds = append(listIds, ref.ListId)
			}
		}

		for _, listId := range listIds {
			item, err := models.SharedIPItemDAO.FindEnabledItemContainsIP(tx, listId, ipLong)
			if err != nil {
				return nil, err
			}
			if item != nil {
				listName, err := models.SharedIPListDAO.FindIPListName(tx, listId)
				if err != nil {
					return nil, err
				}
				if len(listName) == 0 {
					listName = "黑名单"
				}
				return &pb.CheckHTTPFirewallPolicyIPStatusResponse{
					IsOk:      true,
					Error:     "",
					IsFound:   true,
					IsAllowed: false,
					IpList:    &pb.IPList{Name: listName, Id: listId},
					IpItem: &pb.IPItem{
						Id:         int64(item.Id),
						IpFrom:     item.IpFrom,
						IpTo:       item.IpTo,
						ExpiredAt:  int64(item.ExpiredAt),
						Reason:     item.Reason,
						Type:       item.Type,
						EventLevel: item.EventLevel,
					},
					RegionCountry:  nil,
					RegionProvince: nil,
				}, nil
			}
		}
	}

	// 检查封禁的地区和省份
	info, err := iplibrary.SharedLibrary.Lookup(req.Ip)
	if err != nil {
		return nil, err
	}
	if info != nil {
		if firewallPolicy.Inbound != nil &&
			firewallPolicy.Inbound.IsOn &&
			firewallPolicy.Inbound.Region != nil &&
			firewallPolicy.Inbound.Region.IsOn {
			// 检查封禁的地区
			countryId, err := regions.SharedRegionCountryDAO.FindCountryIdWithNameCacheable(tx, info.Country)
			if err != nil {
				return nil, err
			}
			if countryId > 0 && lists.ContainsInt64(firewallPolicy.Inbound.Region.DenyCountryIds, countryId) {
				return &pb.CheckHTTPFirewallPolicyIPStatusResponse{
					IsOk:      true,
					Error:     "",
					IsFound:   true,
					IsAllowed: false,
					IpList:    nil,
					IpItem:    nil,
					RegionCountry: &pb.RegionCountry{
						Id:   countryId,
						Name: info.Country,
					},
					RegionProvince: nil,
				}, nil
			}

			// 检查封禁的省份
			if countryId > 0 {
				provinceId, err := regions.SharedRegionProvinceDAO.FindProvinceIdWithNameCacheable(tx, countryId, info.Province)
				if err != nil {
					return nil, err
				}
				if provinceId > 0 && lists.ContainsInt64(firewallPolicy.Inbound.Region.DenyProvinceIds, provinceId) {
					return &pb.CheckHTTPFirewallPolicyIPStatusResponse{
						IsOk:      true,
						Error:     "",
						IsFound:   true,
						IsAllowed: false,
						IpList:    nil,
						IpItem:    nil,
						RegionCountry: &pb.RegionCountry{
							Id:   countryId,
							Name: info.Country,
						},
						RegionProvince: &pb.RegionProvince{
							Id:   provinceId,
							Name: info.Province,
						},
					}, nil
				}
			}
		}
	}

	return &pb.CheckHTTPFirewallPolicyIPStatusResponse{
		IsOk:           true,
		Error:          "",
		IsFound:        false,
		IsAllowed:      false,
		IpList:         nil,
		IpItem:         nil,
		RegionCountry:  nil,
		RegionProvince: nil,
	}, nil
}
