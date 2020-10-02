package services

import (
	"context"
	"encoding/json"
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

// 创建缓存策略
func (this *HTTPCachePolicyService) CreateHTTPCachePolicy(ctx context.Context, req *pb.CreateHTTPCachePolicyRequest) (*pb.CreateHTTPCachePolicyResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policyId, err := models.SharedHTTPCachePolicyDAO.CreateCachePolicy(req.IsOn, req.Name, req.Description, req.CapacityJSON, req.MaxKeys, req.MaxSizeJSON, req.Type, req.OptionsJSON)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPCachePolicyResponse{CachePolicyId: policyId}, nil
}

// 修改缓存策略
func (this *HTTPCachePolicyService) UpdateHTTPCachePolicy(ctx context.Context, req *pb.UpdateHTTPCachePolicyRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPCachePolicyDAO.UpdateCachePolicy(req.CachePolicyId, req.IsOn, req.Name, req.Description, req.CapacityJSON, req.MaxKeys, req.MaxSizeJSON, req.Type, req.OptionsJSON)
	if err != nil {
		return nil, err
	}

	return rpcutils.RPCUpdateSuccess()
}

// 删除缓存策略
func (this *HTTPCachePolicyService) DeleteHTTPCachePolicy(ctx context.Context, req *pb.DeleteHTTPCachePolicyRequest) (*pb.RPCDeleteSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPCachePolicyDAO.DisableHTTPCachePolicy(req.CachePolicyId)
	if err != nil {
		return nil, err
	}

	return rpcutils.RPCDeleteSuccess()
}

// 计算缓存策略数量
func (this *HTTPCachePolicyService) CountAllEnabledHTTPCachePolicies(ctx context.Context, req *pb.CountAllEnabledHTTPCachePoliciesRequest) (*pb.CountAllEnabledHTTPCachePoliciesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedHTTPCachePolicyDAO.CountAllEnabledHTTPCachePolicies()
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledHTTPCachePoliciesResponse{Count: count}, nil
}

// 列出单页的缓存策略
func (this *HTTPCachePolicyService) ListEnabledHTTPCachePolicies(ctx context.Context, req *pb.ListEnabledHTTPCachePoliciesRequest) (*pb.ListEnabledHTTPCachePoliciesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	cachePolicies, err := models.SharedHTTPCachePolicyDAO.ListEnabledHTTPCachePolicies(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	cachePoliciesJSON, err := json.Marshal(cachePolicies)
	if err != nil {
		return nil, err
	}
	return &pb.ListEnabledHTTPCachePoliciesResponse{CachePoliciesJSON: cachePoliciesJSON}, nil
}

// 查找单个缓存策略配置
func (this *HTTPCachePolicyService) FindEnabledHTTPCachePolicyConfig(ctx context.Context, req *pb.FindEnabledHTTPCachePolicyConfigRequest) (*pb.FindEnabledHTTPCachePolicyConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	cachePolicy, err := models.SharedHTTPCachePolicyDAO.ComposeCachePolicy(req.CachePolicyId)
	if err != nil {
		return nil, err
	}
	cachePolicyJSON, err := json.Marshal(cachePolicy)
	return &pb.FindEnabledHTTPCachePolicyConfigResponse{CachePolicyJSON: cachePolicyJSON}, nil
}
