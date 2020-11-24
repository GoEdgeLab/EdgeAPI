package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
)

// SSL证书相关服务
type SSLCertService struct {
	BaseService
}

// 创建Cert
func (this *SSLCertService) CreateSSLCert(ctx context.Context, req *pb.CreateSSLCertRequest) (*pb.CreateSSLCertResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	certId, err := models.SharedSSLCertDAO.CreateCert(req.IsOn, req.Name, req.Description, req.ServerName, req.IsCA, req.CertData, req.KeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return &pb.CreateSSLCertResponse{CertId: certId}, nil
}

// 修改Cert
func (this *SSLCertService) UpdateSSLCert(ctx context.Context, req *pb.UpdateSSLCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedSSLCertDAO.UpdateCert(req.CertId, req.IsOn, req.Name, req.Description, req.ServerName, req.IsCA, req.CertData, req.KeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找证书配置
func (this *SSLCertService) FindEnabledSSLCertConfig(ctx context.Context, req *pb.FindEnabledSSLCertConfigRequest) (*pb.FindEnabledSSLCertConfigResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	config, err := models.SharedSSLCertDAO.ComposeCertConfig(req.CertId)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledSSLCertConfigResponse{CertJSON: configJSON}, nil
}

// 删除证书
func (this *SSLCertService) DeleteSSLCert(ctx context.Context, req *pb.DeleteSSLCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedSSLCertDAO.DisableSSLCert(req.CertId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 计算匹配的Cert数量
func (this *SSLCertService) CountSSLCerts(ctx context.Context, req *pb.CountSSLCertRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedSSLCertDAO.CountCerts(req.IsCA, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword)
	if err != nil {
		return nil, err
	}

	return &pb.RPCCountResponse{
		Count: count,
	}, nil
}

// 列出单页匹配的Cert
func (this *SSLCertService) ListSSLCerts(ctx context.Context, req *pb.ListSSLCertsRequest) (*pb.ListSSLCertsResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	certIds, err := models.SharedSSLCertDAO.ListCertIds(req.IsCA, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	certConfigs := []*sslconfigs.SSLCertConfig{}
	for _, certId := range certIds {
		certConfig, err := models.SharedSSLCertDAO.ComposeCertConfig(certId)
		if err != nil {
			return nil, err
		}

		// 这里不需要数据内容
		certConfig.CertData = nil
		certConfig.KeyData = nil

		certConfigs = append(certConfigs, certConfig)
	}
	certConfigsJSON, err := json.Marshal(certConfigs)
	if err != nil {
		return nil, err
	}
	return &pb.ListSSLCertsResponse{CertsJSON: certConfigsJSON}, nil
}
