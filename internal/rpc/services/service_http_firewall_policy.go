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

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(policyId, inboundConfigJSON, outboundConfigJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPFirewallPolicyResponse{FirewallPolicyId: policyId}, nil
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
	firewallPolicy, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(req.FirewallPolicyId)
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

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicy(req.FirewallPolicyId, req.IsOn, req.Name, req.Description, inboundConfigJSON, outboundConfigJSON, req.BlockOptionsJSON)
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

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInboundAndOutbound(req.FirewallPolicyId, req.InboundJSON, req.OutboundJSON)
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

	err = models.SharedHTTPFirewallPolicyDAO.UpdateFirewallPolicyInbound(req.FirewallPolicyId, req.InboundJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 计算可用的防火墙策略数量
func (this *HTTPFirewallPolicyService) CountAllEnabledFirewallPolicies(ctx context.Context, req *pb.CountAllEnabledFirewallPoliciesRequest) (*pb.RPCCountResponse, error) {
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
func (this *HTTPFirewallPolicyService) ListEnabledFirewallPolicies(ctx context.Context, req *pb.ListEnabledFirewallPoliciesRequest) (*pb.ListEnabledFirewallPoliciesResponse, error) {
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

	return &pb.ListEnabledFirewallPoliciesResponse{FirewallPolicies: result}, nil
}

// 删除某个防火墙策略
func (this *HTTPFirewallPolicyService) DeleteFirewallPolicy(ctx context.Context, req *pb.DeleteFirewallPolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallPolicyDAO.DisableHTTPFirewallPolicy(req.FirewallPolicyId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找单个防火墙配置
func (this *HTTPFirewallPolicyService) FindEnabledFirewallPolicyConfig(ctx context.Context, req *pb.FindEnabledFirewallPolicyConfigRequest) (*pb.FindEnabledFirewallPolicyConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedHTTPFirewallPolicyDAO.ComposeFirewallPolicy(req.FirewallPolicyId)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return &pb.FindEnabledFirewallPolicyConfigResponse{FirewallPolicyJSON: nil}, nil
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledFirewallPolicyConfigResponse{FirewallPolicyJSON: configJSON}, nil
}

// 获取防火墙的基本信息
func (this *HTTPFirewallPolicyService) FindEnabledFirewallPolicy(ctx context.Context, req *pb.FindEnabledFirewallPolicyRequest) (*pb.FindEnabledFirewallPolicyResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policy, err := models.SharedHTTPFirewallPolicyDAO.FindEnabledHTTPFirewallPolicy(req.FirewallPolicyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return &pb.FindEnabledFirewallPolicyResponse{FirewallPolicy: nil}, nil
	}
	return &pb.FindEnabledFirewallPolicyResponse{FirewallPolicy: &pb.HTTPFirewallPolicy{
		Id:           int64(policy.Id),
		Name:         policy.Name,
		Description:  policy.Description,
		IsOn:         policy.IsOn == 1,
		InboundJSON:  []byte(policy.Inbound),
		OutboundJSON: []byte(policy.Outbound),
	}}, nil
}
