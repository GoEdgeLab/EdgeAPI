package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
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
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	certId, err := models.SharedSSLCertDAO.CreateCert(adminId, userId, req.IsOn, req.Name, req.Description, req.ServerName, req.IsCA, req.CertData, req.KeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return &pb.CreateSSLCertResponse{SslCertId: certId}, nil
}

// 修改Cert
func (this *SSLCertService) UpdateSSLCert(ctx context.Context, req *pb.UpdateSSLCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedSSLCertDAO.UpdateCert(req.SslCertId, req.IsOn, req.Name, req.Description, req.ServerName, req.IsCA, req.CertData, req.KeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找证书配置
func (this *SSLCertService) FindEnabledSSLCertConfig(ctx context.Context, req *pb.FindEnabledSSLCertConfigRequest) (*pb.FindEnabledSSLCertConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedSSLCertDAO.ComposeCertConfig(req.SslCertId)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledSSLCertConfigResponse{SslCertJSON: configJSON}, nil
}

// 删除证书
func (this *SSLCertService) DeleteSSLCert(ctx context.Context, req *pb.DeleteSSLCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedSSLCertDAO.DisableSSLCert(req.SslCertId)
	if err != nil {
		return nil, err
	}

	// 停止相关ACME任务
	err = models.SharedACMETaskDAO.DisableAllTasksWithCertId(req.SslCertId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 计算匹配的Cert数量
func (this *SSLCertService) CountSSLCerts(ctx context.Context, req *pb.CountSSLCertRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedSSLCertDAO.CountCerts(req.IsCA, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, req.UserId)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// 列出单页匹配的Cert
func (this *SSLCertService) ListSSLCerts(ctx context.Context, req *pb.ListSSLCertsRequest) (*pb.ListSSLCertsResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	certIds, err := models.SharedSSLCertDAO.ListCertIds(req.IsCA, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, req.UserId, req.Offset, req.Size)
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
	return &pb.ListSSLCertsResponse{SslCertsJSON: certConfigsJSON}, nil
}
