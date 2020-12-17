package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
	"github.com/iwind/TeaGo/lists"
)

// HTTP防火墙（WAF）相关服务
type HTTPFirewallPolicyService struct {
	BaseService
}

// 获取所有可用策略
func (this *HTTPFirewallPolicyService) FindAllEnabledHTTPFirewallPolicies(ctx context.Context, req *pb.FindAllEnabledHTTPFirewallPoliciesRequest) (*pb.FindAllEnabledHTTPFirewallPoliciesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policies, err := models.SharedHTTPFirewallPolicyDAO.FindAllEnabledFirewallPolicies()
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
		})
	}

	return &pb.FindAllEnabledHTTPFirewallPoliciesResponse{FirewallPolicies: result}, nil
}

// 创建防火墙策略
func (this *HTTPFirewallPolicyService) CreateHTTPFirewallPolicy(ctx context.Context, req *pb.CreateHTTPFirewallPolicyRequest) (*pb.CreateHTTPFirewallPolicyResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policyId, err := models.SharedHTTPFirewallPolicyDAO.CreateFirewallPolicy(req.IsOn, req.Name, req.Description, nil, nil)
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

			groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(group)
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

			groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(group)
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

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(policyId, inboundConfigJSON, outboundConfigJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPFirewallPolicyResponse{HttpFirewallPolicyId: policyId}, nil
}

// 修改防火墙策略
func (this *HTTPFirewallPolicyService) UpdateHTTPFirewallPolicy(ctx context.Context, req *pb.UpdateHTTPFirewallPolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	templatePolicy := firewallconfigs.HTTPFirewallTemplate()

	// 已经有的数据
	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(req.HttpFirewallPolicyId)
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
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(g.Id, true)
					if err != nil {
						return nil, err
					}
				} else {
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(g.Id, false)
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
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(g.Id, true)
					if err != nil {
						return nil, err
					}
				} else {
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(g.Id, false)
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

			groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(group)
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

			groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(group)
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

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicy(req.HttpFirewallPolicyId, req.IsOn, req.Name, req.Description, inboundConfigJSON, outboundConfigJSON, req.BlockOptionsJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改分组信息
func (this *HTTPFirewallPolicyService) UpdateHTTPFirewallPolicyGroups(ctx context.Context, req *pb.UpdateHTTPFirewallPolicyGroupsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(req.HttpFirewallPolicyId, req.InboundJSON, req.OutboundJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 修改inbound信息
func (this *HTTPFirewallPolicyService) UpdateHTTPFirewallInboundConfig(ctx context.Context, req *pb.UpdateHTTPFirewallInboundConfigRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInbound(req.HttpFirewallPolicyId, req.InboundJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 计算可用的防火墙策略数量
func (this *HTTPFirewallPolicyService) CountAllEnabledHTTPFirewallPolicies(ctx context.Context, req *pb.CountAllEnabledHTTPFirewallPoliciesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedHTTPFirewallPolicyDAO.CountAllEnabledFirewallPolicies()
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 列出单页的防火墙策略
func (this *HTTPFirewallPolicyService) ListEnabledHTTPFirewallPolicies(ctx context.Context, req *pb.ListEnabledHTTPFirewallPoliciesRequest) (*pb.ListEnabledHTTPFirewallPoliciesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policies, err := models.SharedHTTPFirewallPolicyDAO.ListEnabledFirewallPolicies(req.Offset, req.Size)
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
		})
	}

	return &pb.ListEnabledHTTPFirewallPoliciesResponse{HttpFirewallPolicies: result}, nil
}

// 删除某个防火墙策略
func (this *HTTPFirewallPolicyService) DeleteHTTPFirewallPolicy(ctx context.Context, req *pb.DeleteHTTPFirewallPolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.DisableHTTPFirewallPolicy(req.HttpFirewallPolicyId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找单个防火墙配置
func (this *HTTPFirewallPolicyService) FindEnabledHTTPFirewallPolicyConfig(ctx context.Context, req *pb.FindEnabledHTTPFirewallPolicyConfigRequest) (*pb.FindEnabledHTTPFirewallPolicyConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(req.HttpFirewallPolicyId)
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

// 获取防火墙的基本信息
func (this *HTTPFirewallPolicyService) FindEnabledHTTPFirewallPolicy(ctx context.Context, req *pb.FindEnabledHTTPFirewallPolicyRequest) (*pb.FindEnabledHTTPFirewallPolicyResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policy, err := models.SharedHTTPFirewallPolicyDAO.FindEnabledHTTPFirewallPolicy(req.HttpFirewallPolicyId)
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
	}}, nil
}

// 导入策略数据
func (this *HTTPFirewallPolicyService) ImportHTTPFirewallPolicy(ctx context.Context, req *pb.ImportHTTPFirewallPolicyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 检查权限

	oldConfig, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(req.HttpFirewallPolicyId)
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
			if len(g.Code) > 0 {
				// 对于有代号的，覆盖或者添加
				oldGroup := oldConfig.FindRuleGroupWithCode(g.Code)
				if oldGroup == nil {
					// 新创建分组
					groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(g)
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
						setId, err := models.SharedHTTPFirewallRuleSetDAO.CreateOrUpdateSetFromConfig(set)
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
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(oldGroup.Id, true)
					if err != nil {
						return nil, err
					}
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupSets(oldGroup.Id, setsJSON)
					if err != nil {
						return nil, err
					}
				}
			} else {
				// 没有代号的直接创建
				groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(g)
				if err != nil {
					return nil, err
				}
				oldConfig.Inbound.GroupRefs = append(oldConfig.Inbound.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
					IsOn:    true,
					GroupId: groupId,
				})
			}
		}
	}

	// 出站分组
	if newConfig.Outbound != nil {
		for _, g := range newConfig.Outbound.Groups {
			if len(g.Code) > 0 {
				// 对于有代号的，覆盖或者添加
				oldGroup := oldConfig.FindRuleGroupWithCode(g.Code)
				if oldGroup == nil {
					// 新创建分组
					groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(g)
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
						setId, err := models.SharedHTTPFirewallRuleSetDAO.CreateOrUpdateSetFromConfig(set)
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
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(oldGroup.Id, true)
					if err != nil {
						return nil, err
					}
					err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupSets(oldGroup.Id, setsJSON)
					if err != nil {
						return nil, err
					}
				}
			} else {
				// 没有代号的直接创建
				groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroupFromConfig(g)
				if err != nil {
					return nil, err
				}
				oldConfig.Outbound.GroupRefs = append(oldConfig.Outbound.GroupRefs, &firewallconfigs.HTTPFirewallRuleGroupRef{
					IsOn:    true,
					GroupId: groupId,
				})
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

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(req.HttpFirewallPolicyId, inboundJSON, outboundJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
