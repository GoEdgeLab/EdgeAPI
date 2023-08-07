package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
)

type HTTPCachePolicyService struct {
	BaseService
}

// FindAllEnabledHTTPCachePolicies 获取所有可用策略
func (this *HTTPCachePolicyService) FindAllEnabledHTTPCachePolicies(ctx context.Context, req *pb.FindAllEnabledHTTPCachePoliciesRequest) (*pb.FindAllEnabledHTTPCachePoliciesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	policies, err := models.SharedHTTPCachePolicyDAO.FindAllEnabledCachePolicies(tx)
	if err != nil {
		return nil, err
	}
	result := []*pb.HTTPCachePolicy{}
	for _, p := range policies {
		result = append(result, &pb.HTTPCachePolicy{
			Id:   int64(p.Id),
			Name: p.Name,
			IsOn: p.IsOn,
		})
	}
	return &pb.FindAllEnabledHTTPCachePoliciesResponse{CachePolicies: result}, nil
}

// CreateHTTPCachePolicy 创建缓存策略
func (this *HTTPCachePolicyService) CreateHTTPCachePolicy(ctx context.Context, req *pb.CreateHTTPCachePolicyRequest) (*pb.CreateHTTPCachePolicyResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if req.CapacityJSON != nil {
		req.CapacityJSON, err = utils.JSONDecodeConfig(req.CapacityJSON, &shared.SizeCapacity{})
		if err != nil {
			return nil, err
		}
	}
	if req.MaxSizeJSON != nil {
		req.MaxSizeJSON, err = utils.JSONDecodeConfig(req.MaxSizeJSON, &shared.SizeCapacity{})
		if err != nil {
			return nil, err
		}
	}
	if req.FetchTimeoutJSON != nil {
		req.FetchTimeoutJSON, err = utils.JSONDecodeConfig(req.FetchTimeoutJSON, &shared.TimeDuration{})
		if err != nil {
			return nil, err
		}
	}

	policyId, err := models.SharedHTTPCachePolicyDAO.CreateCachePolicy(tx, req.IsOn, req.Name, req.Description, req.CapacityJSON, req.MaxSizeJSON, req.Type, req.OptionsJSON, req.SyncCompressionCache, req.FetchTimeoutJSON)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPCachePolicyResponse{HttpCachePolicyId: policyId}, nil
}

// UpdateHTTPCachePolicy 修改缓存策略
func (this *HTTPCachePolicyService) UpdateHTTPCachePolicy(ctx context.Context, req *pb.UpdateHTTPCachePolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if req.CapacityJSON != nil {
		req.CapacityJSON, err = utils.JSONDecodeConfig(req.CapacityJSON, &shared.SizeCapacity{})
		if err != nil {
			return nil, err
		}
	}
	if req.MaxSizeJSON != nil {
		req.MaxSizeJSON, err = utils.JSONDecodeConfig(req.MaxSizeJSON, &shared.SizeCapacity{})
		if err != nil {
			return nil, err
		}
	}
	if req.FetchTimeoutJSON != nil {
		req.FetchTimeoutJSON, err = utils.JSONDecodeConfig(req.FetchTimeoutJSON, &shared.TimeDuration{})
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPCachePolicyDAO.UpdateCachePolicy(tx, req.HttpCachePolicyId, req.IsOn, req.Name, req.Description, req.CapacityJSON, req.MaxSizeJSON, req.Type, req.OptionsJSON, req.SyncCompressionCache, req.FetchTimeoutJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteHTTPCachePolicy 删除缓存策略
func (this *HTTPCachePolicyService) DeleteHTTPCachePolicy(ctx context.Context, req *pb.DeleteHTTPCachePolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedHTTPCachePolicyDAO.DisableHTTPCachePolicy(tx, req.HttpCachePolicyId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountAllEnabledHTTPCachePolicies 计算缓存策略数量
func (this *HTTPCachePolicyService) CountAllEnabledHTTPCachePolicies(ctx context.Context, req *pb.CountAllEnabledHTTPCachePoliciesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedHTTPCachePolicyDAO.CountAllEnabledHTTPCachePolicies(tx, req.NodeClusterId, req.Keyword, req.Type)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledHTTPCachePolicies 列出单页的缓存策略
func (this *HTTPCachePolicyService) ListEnabledHTTPCachePolicies(ctx context.Context, req *pb.ListEnabledHTTPCachePoliciesRequest) (*pb.ListEnabledHTTPCachePoliciesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	cachePolicies, err := models.SharedHTTPCachePolicyDAO.ListEnabledHTTPCachePolicies(tx, req.NodeClusterId, req.Keyword, req.Type, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	cachePoliciesJSON, err := json.Marshal(cachePolicies)
	if err != nil {
		return nil, err
	}
	return &pb.ListEnabledHTTPCachePoliciesResponse{HttpCachePoliciesJSON: cachePoliciesJSON}, nil
}

// FindEnabledHTTPCachePolicyConfig 查找单个缓存策略配置
func (this *HTTPCachePolicyService) FindEnabledHTTPCachePolicyConfig(ctx context.Context, req *pb.FindEnabledHTTPCachePolicyConfigRequest) (*pb.FindEnabledHTTPCachePolicyConfigResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	cachePolicy, err := models.SharedHTTPCachePolicyDAO.ComposeCachePolicy(tx, req.HttpCachePolicyId, nil)
	if err != nil {
		return nil, err
	}
	cachePolicyJSON, err := json.Marshal(cachePolicy)
	return &pb.FindEnabledHTTPCachePolicyConfigResponse{HttpCachePolicyJSON: cachePolicyJSON}, nil
}

// FindEnabledHTTPCachePolicy 查找单个缓存策略信息
func (this *HTTPCachePolicyService) FindEnabledHTTPCachePolicy(ctx context.Context, req *pb.FindEnabledHTTPCachePolicyRequest) (*pb.FindEnabledHTTPCachePolicyResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	policy, err := models.SharedHTTPCachePolicyDAO.FindEnabledHTTPCachePolicy(tx, req.HttpCachePolicyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return &pb.FindEnabledHTTPCachePolicyResponse{HttpCachePolicy: nil}, nil
	}
	return &pb.FindEnabledHTTPCachePolicyResponse{HttpCachePolicy: &pb.HTTPCachePolicy{
		Id:           int64(policy.Id),
		Name:         policy.Name,
		IsOn:         policy.IsOn,
		MaxBytesJSON: policy.MaxSize,
	}}, nil
}

// UpdateHTTPCachePolicyRefs 设置缓存策略的默认条件
func (this *HTTPCachePolicyService) UpdateHTTPCachePolicyRefs(ctx context.Context, req *pb.UpdateHTTPCachePolicyRefsRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedHTTPCachePolicyDAO.UpdatePolicyRefs(tx, req.HttpCachePolicyId, req.RefsJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
