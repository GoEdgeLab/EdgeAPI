package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// HTTPFirewallRuleGroupService WAF规则分组相关服务
type HTTPFirewallRuleGroupService struct {
	BaseService
}

// UpdateHTTPFirewallRuleGroupIsOn 设置是否启用分组
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroupIsOn(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupIsOnRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 校验权限
		err = models.SharedHTTPFirewallRuleGroupDAO.CheckUserRuleGroup(nil, userId, req.FirewallRuleGroupId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupIsOn(tx, req.FirewallRuleGroupId, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CreateHTTPFirewallRuleGroup 创建分组
func (this *HTTPFirewallRuleGroupService) CreateHTTPFirewallRuleGroup(ctx context.Context, req *pb.CreateHTTPFirewallRuleGroupRequest) (*pb.CreateHTTPFirewallRuleGroupResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	groupId, err := models.SharedHTTPFirewallRuleGroupDAO.CreateGroup(tx, req.IsOn, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPFirewallRuleGroupResponse{FirewallRuleGroupId: groupId}, nil
}

// UpdateHTTPFirewallRuleGroup 修改分组
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroup(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 校验权限
		err = models.SharedHTTPFirewallRuleGroupDAO.CheckUserRuleGroup(nil, userId, req.FirewallRuleGroupId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroup(tx, req.FirewallRuleGroupId, req.IsOn, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledHTTPFirewallRuleGroupConfig 获取分组配置
func (this *HTTPFirewallRuleGroupService) FindEnabledHTTPFirewallRuleGroupConfig(ctx context.Context, req *pb.FindEnabledHTTPFirewallRuleGroupConfigRequest) (*pb.FindEnabledHTTPFirewallRuleGroupConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 校验权限
		err = models.SharedHTTPFirewallRuleGroupDAO.CheckUserRuleGroup(nil, userId, req.FirewallRuleGroupId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	groupConfig, err := models.SharedHTTPFirewallRuleGroupDAO.ComposeFirewallRuleGroup(tx, req.FirewallRuleGroupId)
	if err != nil {
		return nil, err
	}
	if groupConfig == nil {
		return &pb.FindEnabledHTTPFirewallRuleGroupConfigResponse{FirewallRuleGroupJSON: nil}, nil
	}
	groupConfigJSON, err := json.Marshal(groupConfig)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledHTTPFirewallRuleGroupConfigResponse{FirewallRuleGroupJSON: groupConfigJSON}, nil
}

// FindEnabledHTTPFirewallRuleGroup 获取分组信息
func (this *HTTPFirewallRuleGroupService) FindEnabledHTTPFirewallRuleGroup(ctx context.Context, req *pb.FindEnabledHTTPFirewallRuleGroupRequest) (*pb.FindEnabledHTTPFirewallRuleGroupResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 校验权限
		err = models.SharedHTTPFirewallRuleGroupDAO.CheckUserRuleGroup(nil, userId, req.FirewallRuleGroupId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	group, err := models.SharedHTTPFirewallRuleGroupDAO.FindEnabledHTTPFirewallRuleGroup(tx, req.FirewallRuleGroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindEnabledHTTPFirewallRuleGroupResponse{
			FirewallRuleGroup: nil,
		}, nil
	}

	return &pb.FindEnabledHTTPFirewallRuleGroupResponse{
		FirewallRuleGroup: &pb.HTTPFirewallRuleGroup{
			Id:          int64(group.Id),
			Name:        group.Name,
			IsOn:        group.IsOn == 1,
			Description: group.Description,
			Code:        group.Code,
		},
	}, nil
}

// UpdateHTTPFirewallRuleGroupSets 修改分组的规则集
func (this *HTTPFirewallRuleGroupService) UpdateHTTPFirewallRuleGroupSets(ctx context.Context, req *pb.UpdateHTTPFirewallRuleGroupSetsRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// 校验权限
		err = models.SharedHTTPFirewallRuleGroupDAO.CheckUserRuleGroup(nil, userId, req.FirewallRuleGroupId)
		if err != nil {
			return nil, err
		}
	}
	
	tx := this.NullTx()

	err = models.SharedHTTPFirewallRuleGroupDAO.UpdateGroupSets(tx, req.GetFirewallRuleGroupId(), req.FirewallRuleSetsJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
