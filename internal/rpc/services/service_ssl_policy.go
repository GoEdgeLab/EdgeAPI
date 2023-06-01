package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
)

type SSLPolicyService struct {
	BaseService
}

// CreateSSLPolicy 创建Policy
func (this *SSLPolicyService) CreateSSLPolicy(ctx context.Context, req *pb.CreateSSLPolicyRequest) (*pb.CreateSSLPolicyResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		//  检查证书
		if len(req.SslCertsJSON) > 0 {
			certRefs := []*sslconfigs.SSLCertRef{}
			err = json.Unmarshal(req.SslCertsJSON, &certRefs)
			if err != nil {
				return nil, err
			}
			for _, certRef := range certRefs {
				err = models.SharedSSLCertDAO.CheckUserCert(tx, certRef.CertId, userId)
				if err != nil {
					return nil, err
				}
			}
		}

		// 检查CA证书
		// TODO
	}

	policyId, err := models.SharedSSLPolicyDAO.CreatePolicy(tx, adminId, userId, req.Http2Enabled, req.Http3Enabled, req.MinVersion, req.SslCertsJSON, req.HstsJSON, req.OcspIsOn, req.ClientAuthType, req.ClientCACertsJSON, req.CipherSuitesIsOn, req.CipherSuites)
	if err != nil {
		return nil, err
	}

	return &pb.CreateSSLPolicyResponse{SslPolicyId: policyId}, nil
}

// UpdateSSLPolicy 修改Policy
func (this *SSLPolicyService) UpdateSSLPolicy(ctx context.Context, req *pb.UpdateSSLPolicyRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedSSLPolicyDAO.CheckUserPolicy(tx, userId, req.SslPolicyId)
		if err != nil {
			return nil, errors.New("check ssl policy failed: " + err.Error())
		}
	}

	err = models.SharedSSLPolicyDAO.UpdatePolicy(tx, req.SslPolicyId, req.Http2Enabled, req.Http3Enabled, req.MinVersion, req.SslCertsJSON, req.HstsJSON, req.OcspIsOn, req.ClientAuthType, req.ClientCACertsJSON, req.CipherSuitesIsOn, req.CipherSuites)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledSSLPolicyConfig 查找Policy
func (this *SSLPolicyService) FindEnabledSSLPolicyConfig(ctx context.Context, req *pb.FindEnabledSSLPolicyConfigRequest) (*pb.FindEnabledSSLPolicyConfigResponse, error) {
	// 校验请求
	// 这里不使用validateAdminAndUser()，是因为我们允许用户ID为0的时候也可以调用
	_, _, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	config, err := models.SharedSSLPolicyDAO.ComposePolicyConfig(tx, req.SslPolicyId, req.IgnoreData, nil, nil)
	if err != nil {
		return nil, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledSSLPolicyConfigResponse{SslPolicyJSON: configJSON}, nil
}
