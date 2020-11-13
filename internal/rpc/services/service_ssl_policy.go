package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type SSLPolicyService struct {
}

// 创建Policy
func (this *SSLPolicyService) CreateSSLPolicy(ctx context.Context, req *pb.CreateSSLPolicyRequest) (*pb.CreateSSLPolicyResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	policyId, err := models.SharedSSLPolicyDAO.CreatePolicy(req.Http2Enabled, req.MinVersion, req.CertsJSON, req.HstsJSON, req.ClientAuthType, req.ClientCACertsJSON, req.CipherSuitesIsOn, req.CipherSuites)
	if err != nil {
		return nil, err
	}

	return &pb.CreateSSLPolicyResponse{SslPolicyId: policyId}, nil
}

// 修改Policy
func (this *SSLPolicyService) UpdateSSLPolicy(ctx context.Context, req *pb.UpdateSSLPolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedSSLPolicyDAO.UpdatePolicy(req.SslPolicyId, req.Http2Enabled, req.MinVersion, req.CertsJSON, req.HstsJSON, req.ClientAuthType, req.ClientCACertsJSON, req.CipherSuitesIsOn, req.CipherSuites)
	if err != nil {
		return nil, err
	}

	return rpcutils.Success()
}

// 查找Policy
func (this *SSLPolicyService) FindEnabledSSLPolicyConfig(ctx context.Context, req *pb.FindEnabledSSLPolicyConfigRequest) (*pb.FindEnabledSSLPolicyConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedSSLPolicyDAO.ComposePolicyConfig(req.SslPolicyId)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledSSLPolicyConfigResponse{SslPolicyJSON: configJSON}, nil
}
