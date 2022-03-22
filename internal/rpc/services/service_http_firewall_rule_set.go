package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/firewallconfigs"
)

// HTTPFirewallRuleSetService 规则集相关服务
type HTTPFirewallRuleSetService struct {
	BaseService
}

// CreateOrUpdateHTTPFirewallRuleSetFromConfig 根据配置创建规则集
func (this *HTTPFirewallRuleSetService) CreateOrUpdateHTTPFirewallRuleSetFromConfig(ctx context.Context, req *pb.CreateOrUpdateHTTPFirewallRuleSetFromConfigRequest) (*pb.CreateOrUpdateHTTPFirewallRuleSetFromConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	setConfig := &firewallconfigs.HTTPFirewallRuleSet{}
	err = json.Unmarshal(req.FirewallRuleSetConfigJSON, setConfig)
	if err != nil {
		return nil, err
	}

	if userId > 0 && setConfig.Id > 0 {
		err = models.SharedHTTPFirewallRuleSetDAO.CheckUserRuleSet(nil, userId, setConfig.Id)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	setId, err := models.SharedHTTPFirewallRuleSetDAO.CreateOrUpdateSetFromConfig(tx, setConfig)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrUpdateHTTPFirewallRuleSetFromConfigResponse{FirewallRuleSetId: setId}, nil
}

// UpdateHTTPFirewallRuleSetIsOn 修改是否开启
func (this *HTTPFirewallRuleSetService) UpdateHTTPFirewallRuleSetIsOn(ctx context.Context, req *pb.UpdateHTTPFirewallRuleSetIsOnRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPFirewallRuleSetDAO.CheckUserRuleSet(nil, userId, req.FirewallRuleSetId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	err = models.SharedHTTPFirewallRuleSetDAO.UpdateRuleSetIsOn(tx, req.FirewallRuleSetId, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledHTTPFirewallRuleSetConfig 查找规则集配置
func (this *HTTPFirewallRuleSetService) FindEnabledHTTPFirewallRuleSetConfig(ctx context.Context, req *pb.FindEnabledHTTPFirewallRuleSetConfigRequest) (*pb.FindEnabledHTTPFirewallRuleSetConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPFirewallRuleSetDAO.CheckUserRuleSet(nil, userId, req.FirewallRuleSetId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	config, err := models.SharedHTTPFirewallRuleSetDAO.ComposeFirewallRuleSet(tx, req.FirewallRuleSetId)
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

// FindEnabledHTTPFirewallRuleSet 查找规则集
func (this *HTTPFirewallRuleSetService) FindEnabledHTTPFirewallRuleSet(ctx context.Context, req *pb.FindEnabledHTTPFirewallRuleSetRequest) (*pb.FindEnabledHTTPFirewallRuleSetResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedHTTPFirewallRuleSetDAO.CheckUserRuleSet(nil, userId, req.FirewallRuleSetId)
		if err != nil {
			return nil, err
		}
	}

	tx := this.NullTx()

	set, err := models.SharedHTTPFirewallRuleSetDAO.FindEnabledHTTPFirewallRuleSet(tx, req.FirewallRuleSetId)
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
			IsOn:        set.IsOn,
			Description: set.Description,
			Code:        set.Code,
		},
	}, nil
}
