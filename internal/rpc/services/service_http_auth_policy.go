// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// HTTPAuthPolicyService 服务认证策略服务
type HTTPAuthPolicyService struct {
	BaseService
}

// CreateHTTPAuthPolicy 创建策略
func (this *HTTPAuthPolicyService) CreateHTTPAuthPolicy(ctx context.Context, req *pb.CreateHTTPAuthPolicyRequest) (*pb.CreateHTTPAuthPolicyResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	policyId, err := models.SharedHTTPAuthPolicyDAO.CreateHTTPAuthPolicy(tx, req.Name, req.Type, req.ParamsJSON)
	if err != nil {
		return nil, err
	}
	return &pb.CreateHTTPAuthPolicyResponse{HttpAuthPolicyId: policyId}, nil
}

// UpdateHTTPAuthPolicy 修改策略
func (this *HTTPAuthPolicyService) UpdateHTTPAuthPolicy(ctx context.Context, req *pb.UpdateHTTPAuthPolicyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedHTTPAuthPolicyDAO.UpdateHTTPAuthPolicy(tx, req.HttpAuthPolicyId, req.Name, req.ParamsJSON, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledHTTPAuthPolicy 查找策略信息
func (this *HTTPAuthPolicyService) FindEnabledHTTPAuthPolicy(ctx context.Context, req *pb.FindEnabledHTTPAuthPolicyRequest) (*pb.FindEnabledHTTPAuthPolicyResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	policy, err := models.SharedHTTPAuthPolicyDAO.FindEnabledHTTPAuthPolicy(tx, req.HttpAuthPolicyId)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return &pb.FindEnabledHTTPAuthPolicyResponse{HttpAuthPolicy: nil}, nil
	}

	return &pb.FindEnabledHTTPAuthPolicyResponse{HttpAuthPolicy: &pb.HTTPAuthPolicy{
		Id:         int64(policy.Id),
		IsOn:       policy.IsOn,
		Name:       policy.Name,
		Type:       policy.Type,
		ParamsJSON: policy.Params,
	}}, nil
}
