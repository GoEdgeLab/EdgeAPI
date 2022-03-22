package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/accesslogs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPAccessLogPolicyService struct {
	BaseService
}

// CountAllEnabledHTTPAccessLogPolicies 计算访问日志策略数量
func (this *HTTPAccessLogPolicyService) CountAllEnabledHTTPAccessLogPolicies(ctx context.Context, req *pb.CountAllEnabledHTTPAccessLogPoliciesRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedHTTPAccessLogPolicyDAO.CountAllEnabledPolicies(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledHTTPAccessLogPolicies 列出单页访问日志策略
func (this *HTTPAccessLogPolicyService) ListEnabledHTTPAccessLogPolicies(ctx context.Context, req *pb.ListEnabledHTTPAccessLogPoliciesRequest) (*pb.ListEnabledHTTPAccessLogPoliciesResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	policies, err := models.SharedHTTPAccessLogPolicyDAO.ListEnabledPolicies(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbPolicies = []*pb.HTTPAccessLogPolicy{}
	for _, policy := range policies {
		pbPolicies = append(pbPolicies, &pb.HTTPAccessLogPolicy{
			Id:          int64(policy.Id),
			Name:        policy.Name,
			IsOn:        policy.IsOn,
			Type:        policy.Type,
			OptionsJSON: policy.Options,
			CondsJSON:   policy.Conds,
			IsPublic:    policy.IsPublic == 1,
		})
	}
	return &pb.ListEnabledHTTPAccessLogPoliciesResponse{HttpAccessLogPolicies: pbPolicies}, nil
}

// CreateHTTPAccessLogPolicy 创建访问日志策略
func (this *HTTPAccessLogPolicyService) CreateHTTPAccessLogPolicy(ctx context.Context, req *pb.CreateHTTPAccessLogPolicyRequest) (*pb.CreateHTTPAccessLogPolicyResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 取消别的Public
	if req.IsPublic {
		err = models.SharedHTTPAccessLogPolicyDAO.CancelAllPublicPolicies(tx)
		if err != nil {
			return nil, err
		}
	}

	// 创建
	policyId, err := models.SharedHTTPAccessLogPolicyDAO.CreatePolicy(tx, req.Name, req.Type, req.OptionsJSON, req.CondsJSON, req.IsPublic)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPAccessLogPolicyResponse{HttpAccessLogPolicyId: policyId}, nil
}

// UpdateHTTPAccessLogPolicy 修改访问日志策略
func (this *HTTPAccessLogPolicyService) UpdateHTTPAccessLogPolicy(ctx context.Context, req *pb.UpdateHTTPAccessLogPolicyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 取消别的Public
	if req.IsPublic {
		err = models.SharedHTTPAccessLogPolicyDAO.CancelAllPublicPolicies(tx)
		if err != nil {
			return nil, err
		}
	}

	// 保存修改
	err = models.SharedHTTPAccessLogPolicyDAO.UpdatePolicy(tx, req.HttpAccessLogPolicyId, req.Name, req.OptionsJSON, req.CondsJSON, req.IsPublic, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledHTTPAccessLogPolicy 查找单个访问日志策略
func (this *HTTPAccessLogPolicyService) FindEnabledHTTPAccessLogPolicy(ctx context.Context, req *pb.FindEnabledHTTPAccessLogPolicyRequest) (*pb.FindEnabledHTTPAccessLogPolicyResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	policy, err := models.SharedHTTPAccessLogPolicyDAO.FindEnabledHTTPAccessLogPolicy(tx, req.HttpAccessLogPolicyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return &pb.FindEnabledHTTPAccessLogPolicyResponse{HttpAccessLogPolicy: nil}, nil
	}
	return &pb.FindEnabledHTTPAccessLogPolicyResponse{HttpAccessLogPolicy: &pb.HTTPAccessLogPolicy{
		Id:          int64(policy.Id),
		Name:        policy.Name,
		IsOn:        policy.IsOn,
		Type:        policy.Type,
		OptionsJSON: policy.Options,
		CondsJSON:   policy.Conds,
		IsPublic:    policy.IsPublic == 1,
	}}, nil
}

// DeleteHTTPAccessLogPolicy 删除访问日志策略
func (this *HTTPAccessLogPolicyService) DeleteHTTPAccessLogPolicy(ctx context.Context, req *pb.DeleteHTTPAccessLogPolicyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedHTTPAccessLogPolicyDAO.DisableHTTPAccessLogPolicy(tx, req.HttpAccessLogPolicyId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// WriteHTTPAccessLogPolicy 测试写入某个访问日志策略
func (this *HTTPAccessLogPolicyService) WriteHTTPAccessLogPolicy(ctx context.Context, req *pb.WriteHTTPAccessLogPolicyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	err = accesslogs.SharedStorageManager.Write(req.HttpAccessLogPolicyId, []*pb.HTTPAccessLog{req.HttpAccessLog})
	if err != nil {
		return nil, err
	}
	return this.Success()
}
