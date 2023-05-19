package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
)

type HTTPHeaderPolicyService struct {
	BaseService
}

// FindEnabledHTTPHeaderPolicyConfig 查找策略配置
func (this *HTTPHeaderPolicyService) FindEnabledHTTPHeaderPolicyConfig(ctx context.Context, req *pb.FindEnabledHTTPHeaderPolicyConfigRequest) (*pb.FindEnabledHTTPHeaderPolicyConfigResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPHeaderPolicyDAO.CheckUserHeaderPolicy(tx, userId, req.HttpHeaderPolicyId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(tx, req.HttpHeaderPolicyId)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledHTTPHeaderPolicyConfigResponse{HttpHeaderPolicyJSON: configData}, nil
}

// CreateHTTPHeaderPolicy 创建策略
func (this *HTTPHeaderPolicyService) CreateHTTPHeaderPolicy(ctx context.Context, req *pb.CreateHTTPHeaderPolicyRequest) (*pb.CreateHTTPHeaderPolicyResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	headerPolicyId, err := models.SharedHTTPHeaderPolicyDAO.CreateHeaderPolicy(tx)
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPHeaderPolicyResponse{HttpHeaderPolicyId: headerPolicyId}, nil
}

// UpdateHTTPHeaderPolicyAddingHeaders 修改AddHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyAddingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyAddingHeadersRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPHeaderPolicyDAO.CheckUserHeaderPolicy(tx, userId, req.HttpHeaderPolicyId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateAddingHeaders(tx, req.HttpHeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicySettingHeaders 修改SetHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicySettingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicySettingHeadersRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPHeaderPolicyDAO.CheckUserHeaderPolicy(tx, userId, req.HttpHeaderPolicyId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateSettingHeaders(tx, req.HttpHeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicyAddingTrailers 修改AddTrailers
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyAddingTrailers(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyAddingTrailersRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPHeaderPolicyDAO.CheckUserHeaderPolicy(tx, userId, req.HttpHeaderPolicyId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateAddingTrailers(tx, req.HttpHeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicyReplacingHeaders 修改ReplaceHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyReplacingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyReplacingHeadersRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPHeaderPolicyDAO.CheckUserHeaderPolicy(tx, userId, req.HttpHeaderPolicyId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateReplacingHeaders(tx, req.HttpHeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicyDeletingHeaders 修改删除的Headers
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyDeletingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyDeletingHeadersRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPHeaderPolicyDAO.CheckUserHeaderPolicy(tx, userId, req.HttpHeaderPolicyId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateDeletingHeaders(tx, req.HttpHeaderPolicyId, req.HeaderNames)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicyCORS 修改策略CORS设置
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyCORS(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyCORSRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPHeaderPolicyDAO.CheckUserHeaderPolicy(tx, userId, req.HttpHeaderPolicyId)
		if err != nil {
			return nil, err
		}
	}

	var corsConfig = shared.NewHTTPCORSHeaderConfig()
	err = json.Unmarshal(req.CorsJSON, corsConfig)
	if err != nil {
		return nil, err
	}
	err = corsConfig.Init()
	if err != nil {
		return nil, errors.New("validate CORS config failed: " + err.Error())
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateHeaderPolicyCORS(tx, req.HttpHeaderPolicyId, corsConfig)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// UpdateHTTPHeaderPolicyNonStandardHeaders 修改非标的Headers
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyNonStandardHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyNonStandardHeadersRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = models.SharedHTTPHeaderPolicyDAO.CheckUserHeaderPolicy(tx, userId, req.HttpHeaderPolicyId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateNonStandardHeaders(tx, req.HttpHeaderPolicyId, req.HeaderNames)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
