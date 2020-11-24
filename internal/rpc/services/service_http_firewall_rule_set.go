package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
)

// 规则集相关服务
type HTTPFirewallRuleSetService struct {
	BaseService
}

// 根据配置创建规则集
func (this *HTTPFirewallRuleSetService) CreateOrUpdateHTTPFirewallRuleSetFromConfig(ctx context.Context, req *pb.CreateOrUpdateHTTPFirewallRuleSetFromConfigRequest) (*pb.CreateOrUpdateHTTPFirewallRuleSetFromConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	setConfig := &firewallconfigs.HTTPFirewallRuleSet{}
	err = json.Unmarshal(req.FirewallRuleSetConfigJSON, setConfig)
	if err != nil {
		return nil, err
	}

	setId, err := models.SharedHTTPFirewallRuleSetDAO.CreateOrUpdateSetFromConfig(setConfig)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrUpdateHTTPFirewallRuleSetFromConfigResponse{FirewallRuleSetId: setId}, nil
}

// 修改是否开启
func (this *HTTPFirewallRuleSetService) UpdateHTTPFirewallRuleSetIsOn(ctx context.Context, req *pb.UpdateHTTPFirewallRuleSetIsOnRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallRuleSetDAO.UpdateRuleSetIsOn(req.FirewallRuleSetId, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找规则集配置
func (this *HTTPFirewallRuleSetService) FindEnabledHTTPFirewallRuleSetConfig(ctx context.Context, req *pb.FindEnabledHTTPFirewallRuleSetConfigRequest) (*pb.FindEnabledHTTPFirewallRuleSetConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedHTTPFirewallRuleSetDAO.ComposeFirewallRuleSet(req.FirewallRuleSetId)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return &pb.FindEnabledHTTPFirewallRuleSetConfigResponse{FirewallRuleSetJSON: nil}, nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledHTTPFirewallRuleSetConfigResponse{FirewallRuleSetJSON: configJSON}, nil
}

// 查找规则集
func (this *HTTPFirewallRuleSetService) FindEnabledHTTPFirewallRuleSet(ctx context.Context, req *pb.FindEnabledHTTPFirewallRuleSetRequest) (*pb.FindEnabledHTTPFirewallRuleSetResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	set, err := models.SharedHTTPFirewallRuleSetDAO.FindEnabledHTTPFirewallRuleSet(req.FirewallRuleSetId)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return &pb.FindEnabledHTTPFirewallRuleSetResponse{
			FirewallRuleSet: nil,
		}, nil
	}

	return &pb.FindEnabledHTTPFirewallRuleSetResponse{
		FirewallRuleSet: &pb.HTTPFirewallRuleSet{
			Id:          int64(set.Id),
			Name:        set.Name,
			IsOn:        set.IsOn == 1,
			Description: set.Description,
			Code:        set.Code,
		},
	}, nil
}
