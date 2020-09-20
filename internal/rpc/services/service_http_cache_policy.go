package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPCachePolicyService struct {
}

// 获取所有可用策略
func (this *HTTPCachePolicyService) FindAllEnabledHTTPCachePolicies(ctx context.Context, req *pb.FindAllEnabledHTTPCachePoliciesRequest) (*pb.FindAllEnabledHTTPCachePoliciesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policies, err := models.SharedHTTPCachePolicyDAO.FindAllEnabledCachePolicies()
	if err != nil {
		return nil, err
	}
	result := []*pb.HTTPCachePolicy{}
	for _, p := range policies {
		result = append(result, &pb.HTTPCachePolicy{
			Id:   int64(p.Id),
			Name: p.Name,
			IsOn: p.IsOn == 1,
		})
	}
	return &pb.FindAllEnabledHTTPCachePoliciesResponse{CachePolicies: result}, nil
}
