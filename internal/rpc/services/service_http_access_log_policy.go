package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPAccessLogPolicyService struct {
	BaseService
}

// FindAllEnabledHTTPAccessLogPolicies 获取所有可用策略
func (this *HTTPAccessLogPolicyService) FindAllEnabledHTTPAccessLogPolicies(ctx context.Context, req *pb.FindAllEnabledHTTPAccessLogPoliciesRequest) (*pb.FindAllEnabledHTTPAccessLogPoliciesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	policies, err := models.SharedHTTPAccessLogPolicyDAO.FindAllEnabledAccessLogPolicies(tx)
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
			CondsJSON:   []byte(policy.Conds),
		})
	}

	return &pb.FindAllEnabledHTTPAccessLogPoliciesResponse{AccessLogPolicies: result}, nil
}
