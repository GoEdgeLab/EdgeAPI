package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type HTTPHeaderPolicyService struct {
}

// 查找策略配置
func (this *HTTPHeaderPolicyService) FindEnabledHTTPHeaderPolicyConfig(ctx context.Context, req *pb.FindEnabledHTTPHeaderPolicyConfigRequest) (*pb.FindEnabledHTTPHeaderPolicyConfigResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedHTTPHeaderPolicyDAO.ComposeHeaderPolicyConfig(req.HeaderPolicyId)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledHTTPHeaderPolicyConfigResponse{HeaderPolicyJSON: configData}, nil
}

// 创建策略
func (this *HTTPHeaderPolicyService) CreateHTTPHeaderPolicy(ctx context.Context, req *pb.CreateHTTPHeaderPolicyRequest) (*pb.CreateHTTPHeaderPolicyResponse, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	headerPolicyId, err := models.SharedHTTPHeaderPolicyDAO.CreateHeaderPolicy()
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPHeaderPolicyResponse{HeaderPolicyId: headerPolicyId}, nil
}

// 修改AddHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyAddingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyAddingHeadersRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateAddingHeaders(req.HeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改SetHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicySettingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicySettingHeadersRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateSettingHeaders(req.HeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改AddTrailers
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyAddingTrailers(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyAddingTrailersRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateAddingTrailers(req.HeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改ReplaceHeaders
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyReplacingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyReplacingHeadersRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateReplacingHeaders(req.HeaderPolicyId, req.HeadersJSON)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}

// 修改删除的Headers
func (this *HTTPHeaderPolicyService) UpdateHTTPHeaderPolicyDeletingHeaders(ctx context.Context, req *pb.UpdateHTTPHeaderPolicyDeletingHeadersRequest) (*pb.RPCUpdateSuccess, error) {
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedHTTPHeaderPolicyDAO.UpdateDeletingHeaders(req.HeaderPolicyId, req.HeaderNames)
	if err != nil {
		return nil, err
	}

	return &pb.RPCUpdateSuccess{}, nil
}
