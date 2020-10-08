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
func (this *HTTPFirewallRuleSetService) UpdateHTTPFirewallRuleSetIsOn(ctx context.Context, req *pb.UpdateHTTPFirewallRuleSetIsOnRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPFirewallRuleSetDAO.UpdateRuleSetIsOn(req.FirewallRuleSetId, req.IsOn)
	if err != nil {
		return nil, err
	}

	return rpcutils.RPCUpdateSuccess()
}

// 查找规则集配置
func (this *HTTPFirewallRuleSetService) FindHTTPFirewallRuleSetConfig(ctx context.Context, req *pb.FindHTTPFirewallRuleSetConfigRequest) (*pb.FindHTTPFirewallRuleSetConfigResponse, error) {
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
		return &pb.FindHTTPFirewallRuleSetConfigResponse{FirewallRuleSetJSON: nil}, nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindHTTPFirewallRuleSetConfigResponse{FirewallRuleSetJSON: configJSON}, nil
}
