package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// WAF规则分组相关服务
type HTTPFirewallRuleGroupService struct {
}

// 设置是否启用分组
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroupIsOn(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupIsOnRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(req.FirewallRuleGroupId, req.IsOn)
	if err != nil {
		return nil, err
	}

	return rpcutils.RPCUpdateSuccess()
}

// 创建分组
func (this *HTTPFirewallRuleGroupService) CreateHTTPFirewallRuleGroup(ctx context.Context, req *pb.CreateHTTPFirewallRuleGroupRequest) (*pb.CreateHTTPFirewallRuleGroupResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroup(req.IsOn, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPFirewallRuleGroupResponse{FirewallRuleGroupId: groupId}, nil
}

// 修改分组
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroup(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroup(req.FirewallRuleGroupId, req.IsOn, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return rpcutils.RPCUpdateSuccess()
}

// 获取分组配置
func (this *HTTPFirewallRuleGroupService) FindHTTPFirewallRuleGroupConfig(ctx context.Context, req *pb.FindHTTPFirewallRuleGroupConfigRequest) (*pb.FindHTTPFirewallRuleGroupConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	groupConfig, err := models.SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(req.FirewallRuleGroupId)
	if err != nil {
		return nil, err
	}
	if groupConfig == nil {
		return &pb.FindHTTPFirewallRuleGroupConfigResponse{FirewallRuleGroupJSON: nil}, nil
	}
	groupConfigJSON, err := json.Marshal(groupConfig)
	if err != nil {
		return nil, err
	}
	return &pb.FindHTTPFirewallRuleGroupConfigResponse{FirewallRuleGroupJSON: groupConfigJSON}, nil
}

// 修改分组的规则集
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroupSets(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupSetsRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupSets(req.GetFirewallRuleGroupId(), req.FirewallRuleSetsJSON)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}