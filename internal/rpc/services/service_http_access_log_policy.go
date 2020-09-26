package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPAccessLogPolicyService struct {
}

// 获取所有可用策略
func (this *HTTPAccessLogPolicyService) FindAllEnabledHTTPAccessLogPolicies(ctx context.Context, req *pb.FindAllEnabledHTTPAccessLogPoliciesRequest) (*pb.FindAllEnabledHTTPAccessLogPoliciesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policies, err := models.SharedHTTPAccessLogPolicyDAO.FindAllEnabledAccessLogPolicies()
	if err != nil {
		return nil, err
	}

	result := []*pb.HTTPAccessLogPolicy{}
	for _, policy := range policies {
		result = append(result, &pb.HTTPAccessLogPolicy{
			Id:          int64(policy.Id),
			Name:        policy.Name,
			IsOn:        policy.IsOn == 1,
			Type:        policy.Name,
			OptionsJSON: []byte(policy.Options),
			CondsJSON:   []byte(policy.CondGroups),
		})
	}

	return &pb.FindAllEnabledHTTPAccessLogPoliciesResponse{AccessLogPolicies: result}, nil
}
