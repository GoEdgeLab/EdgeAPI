package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPFirewallPolicyService struct {
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
			Id:   int64(p.Id),
			Name: p.Name,
			IsOn: p.IsOn == 1,
		})
	}

	return &pb.FindAllEnabledHTTPFirewallPoliciesResponse{FirewallPolicies: result}, nil
}
