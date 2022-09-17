package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPHeaderPolicyService struct {
	BaseService
}

// FindEnabledHTTPHeaderPolicyConfig 查找策略配置
func (this *HTTPHeaderPolicyService) FindEnabledHTTPHeaderPolicyConfig(ctx context.Context, req *pb.FindEnabledHTTPHeaderPolicyConfigRequest) (*pb.FindEnabledHTTPHeaderPolicyConfigResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查权限

	config, err := models.SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(tx, req.HeaderPolicyId)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledHTTPHeaderPolicyConfigResponse{HeaderPolicyJSON: configData}, nil
}

// CreateHTTPHeaderPolicy 创建策略
func (this *HTTPHeaderPolicyService) CreateHTTPHeaderPolicy(ctx context.Context, req *pb.CreateHTTPHeaderPolicyRequest) (*pb.CreateHTTPHeaderPolicyResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查权限

	headerPolicyId, err := models.SharedHTTPHeaderPolicyDAO.CreateHeaderPolicy(tx)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPHeaderPolicyResponse{HeaderPolicyId: headerPolicyId}, nil
}

// UpdateHTTPHeaderPolicyAddingHeaders 修改AddHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyAddingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyAddingHeadersRequest) (*pb.RPCSuccess, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查权限

	err = models.SharedHTTPHeaderPolicyDAO.UpdateAddingHeaders(tx, req.HeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicySettingHeaders 修改SetHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicySettingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicySettingHeadersRequest) (*pb.RPCSuccess, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查权限

	err = models.SharedHTTPHeaderPolicyDAO.UpdateSettingHeaders(tx, req.HeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicyAddingTrailers 修改AddTrailers
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyAddingTrailers(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyAddingTrailersRequest) (*pb.RPCSuccess, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查权限

	err = models.SharedHTTPHeaderPolicyDAO.UpdateAddingTrailers(tx, req.HeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicyReplacingHeaders 修改ReplaceHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyReplacingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyReplacingHeadersRequest) (*pb.RPCSuccess, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查权限

	err = models.SharedHTTPHeaderPolicyDAO.UpdateReplacingHeaders(tx, req.HeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicyDeletingHeaders 修改删除的Headers
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyDeletingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyDeletingHeadersRequest) (*pb.RPCSuccess, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// TODO 检查权限

	err = models.SharedHTTPHeaderPolicyDAO.UpdateDeletingHeaders(tx, req.HeaderPolicyId, req.HeaderNames)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
