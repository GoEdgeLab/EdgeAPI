package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/sslconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

// SSLCertService SSL证书相关服务
type SSLCertService struct {
	BaseService
}

// CreateSSLCert 创建证书
func (this *SSLCertService) CreateSSLCert(ctx context.Context, req *pb.CreateSSLCertRequest) (*pb.CreateSSLCertResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	// 用户ID
	if adminId > 0 && req.UserId > 0 {
		userId = req.UserId
	}

	var tx = this.NullTx()

	if req.TimeBeginAt < 0 {
		return nil, errors.New("invalid TimeBeginAt")
	}
	if req.TimeEndAt < 0 {
		return nil, errors.New("invalid TimeEndAt")
	}

	certId, err := models.SharedSSLCertDAO.CreateCert(tx, adminId, userId, req.IsOn, req.Name, req.Description, req.ServerName, req.IsCA, req.CertData, req.KeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return &pb.CreateSSLCertResponse{SslCertId: certId}, nil
}

// CreateSSLCerts 创建一组证书
func (this *SSLCertService) CreateSSLCerts(ctx context.Context, req *pb.CreateSSLCertsRequest) (*pb.CreateSSLCertsResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if adminId > 0 {
		if req.UserId > 0 {
			userId = req.UserId
		} else {
			userId = 0
		}
	}

	var certIds = []int64{}
	err = this.RunTx(func(tx *dbs.Tx) error {
		for _, cert := range req.SSLCerts {
			certId, err := models.SharedSSLCertDAO.CreateCert(tx, adminId, userId, cert.IsOn, cert.Name, cert.Description, cert.ServerName, cert.IsCA, cert.CertData, cert.KeyData, cert.TimeBeginAt, cert.TimeEndAt, cert.DnsNames, cert.CommonNames)
			if err != nil {
				return err
			}
			certIds = append(certIds, certId)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateSSLCertsResponse{SslCertIds: certIds}, nil
}

// UpdateSSLCert 修改Cert
func (this *SSLCertService) UpdateSSLCert(ctx context.Context, req *pb.UpdateSSLCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if req.TimeBeginAt < 0 {
		return nil, errors.New("invalid TimeBeginAt")
	}
	if req.TimeEndAt < 0 {
		return nil, errors.New("invalid TimeEndAt")
	}

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(tx, req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedSSLCertDAO.UpdateCert(tx, req.SslCertId, req.IsOn, req.Name, req.Description, req.ServerName, req.IsCA, req.CertData, req.KeyData, req.TimeBeginAt, req.TimeEndAt, req.DnsNames, req.CommonNames)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledSSLCertConfig 查找证书配置
func (this *SSLCertService) FindEnabledSSLCertConfig(ctx context.Context, req *pb.FindEnabledSSLCertConfigRequest) (*pb.FindEnabledSSLCertConfigResponse, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(tx, req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	config, err := models.SharedSSLCertDAO.ComposeCertConfig(tx, req.SslCertId, false, nil, nil)
	if err != nil {
		return nil, err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &pb.FindEnabledSSLCertConfigResponse{SslCertJSON: configJSON}, nil
}

// DeleteSSLCert 删除证书
func (this *SSLCertService) DeleteSSLCert(ctx context.Context, req *pb.DeleteSSLCertRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err := models.SharedSSLCertDAO.CheckUserCert(tx, req.SslCertId, userId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedSSLCertDAO.DisableSSLCert(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}

	// 停止相关ACME任务
	err = acme.SharedACMETaskDAO.DisableAllTasksWithCertId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CountSSLCerts 计算匹配的Cert数量
func (this *SSLCertService) CountSSLCerts(ctx context.Context, req *pb.CountSSLCertRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if adminId > 0 {
		userId = req.UserId
	} else if userId <= 0 {
		return nil, errors.New("invalid user")
	}

	count, err := models.SharedSSLCertDAO.CountCerts(tx, req.IsCA, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, userId, req.Domains, req.UserOnly)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListSSLCerts 列出单页匹配的Cert
func (this *SSLCertService) ListSSLCerts(ctx context.Context, req *pb.ListSSLCertsRequest) (*pb.ListSSLCertsResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if adminId > 0 {
		userId = req.UserId
	} else if userId <= 0 {
		return nil, errors.New("invalid user")
	}

	var tx = this.NullTx()

	certIds, err := models.SharedSSLCertDAO.ListCertIds(tx, req.IsCA, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, userId, req.Domains, req.UserOnly, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var certConfigs = []*sslconfigs.SSLCertConfig{}
	for _, certId := range certIds {
		certConfig, err := models.SharedSSLCertDAO.ComposeCertConfig(tx, certId, false, nil, nil)
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

// CountAllSSLCertsWithOCSPError 计算有OCSP错误的证书数量
func (this *SSLCertService) CountAllSSLCertsWithOCSPError(ctx context.Context, req *pb.CountAllSSLCertsWithOCSPErrorRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedSSLCertDAO.CountAllSSLCertsWithOCSPError(tx, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListSSLCertsWithOCSPError 列出有OCSP错误的证书
func (this *SSLCertService) ListSSLCertsWithOCSPError(ctx context.Context, req *pb.ListSSLCertsWithOCSPErrorRequest) (*pb.ListSSLCertsWithOCSPErrorResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	certs, err := models.SharedSSLCertDAO.ListSSLCertsWithOCSPError(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbCerts = []*pb.SSLCert{}
	for _, cert := range certs {
		pbCerts = append(pbCerts, &pb.SSLCert{
			Id:            int64(cert.Id),
			IsOn:          cert.IsOn,
			Name:          cert.Name,
			TimeBeginAt:   types.Int64(cert.TimeBeginAt),
			TimeEndAt:     types.Int64(cert.TimeEndAt),
			DnsNames:      cert.DecodeDNSNames(),
			CommonNames:   cert.DecodeCommonNames(),
			IsACME:        cert.IsACME,
			AcmeTaskId:    int64(cert.AcmeTaskId),
			Ocsp:          cert.Ocsp,
			OcspIsUpdated: cert.OcspIsUpdated == 1,
			OcspError:     cert.OcspError,
			Description:   cert.Description,
			IsCA:          cert.IsCA,
			ServerName:    cert.ServerName,
			CreatedAt:     int64(cert.CreatedAt),
			UpdatedAt:     int64(cert.UpdatedAt),
		})
	}

	return &pb.ListSSLCertsWithOCSPErrorResponse{
		SslCerts: pbCerts,
	}, nil
}

// IgnoreSSLCertsWithOCSPError 忽略一组OCSP证书错误
func (this *SSLCertService) IgnoreSSLCertsWithOCSPError(ctx context.Context, req *pb.IgnoreSSLCertsWithOCSPErrorRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedSSLCertDAO.IgnoreSSLCertsWithOCSPError(tx, req.SslCertIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ResetSSLCertsWithOCSPError 重置一组证书OCSP错误状态
func (this *SSLCertService) ResetSSLCertsWithOCSPError(ctx context.Context, req *pb.ResetSSLCertsWithOCSPErrorRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedSSLCertDAO.ResetSSLCertsWithOCSPError(tx, req.SslCertIds)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ResetAllSSLCertsWithOCSPError 重置所有证书OCSP错误状态
func (this *SSLCertService) ResetAllSSLCertsWithOCSPError(ctx context.Context, req *pb.ResetAllSSLCertsWithOCSPErrorRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedSSLCertDAO.ResetAllSSLCertsWithOCSPError(tx)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// ListUpdatedSSLCertOCSP 读取证书的OCSP
func (this *SSLCertService) ListUpdatedSSLCertOCSP(ctx context.Context, req *pb.ListUpdatedSSLCertOCSPRequest) (*pb.ListUpdatedSSLCertOCSPResponse, error) {
	_, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	certs, err := models.SharedSSLCertDAO.ListCertOCSPAfterVersion(tx, req.Version, int64(req.Size))
	if err != nil {
		return nil, err
	}

	var result = []*pb.ListUpdatedSSLCertOCSPResponse_SSLCertOCSP{}
	for _, cert := range certs {
		result = append(result, &pb.ListUpdatedSSLCertOCSPResponse_SSLCertOCSP{
			SslCertId: int64(cert.Id),
			Data:      cert.Ocsp,
			ExpiresAt: int64(cert.OcspExpiresAt),
			Version:   int64(cert.OcspUpdatedVersion),
		})
	}

	return &pb.ListUpdatedSSLCertOCSPResponse{
		SslCertOCSP: result,
	}, nil
}

// FindSSLCertUser 查找证书所属用户
func (this *SSLCertService) FindSSLCertUser(ctx context.Context, req *pb.FindSSLCertUserRequest) (*pb.FindSSLCertUserResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userId, err := models.SharedSSLCertDAO.FindCertUserId(tx, req.SslCertId)
	if err != nil {
		return nil, err
	}
	if userId <= 0 {
		return &pb.FindSSLCertUserResponse{User: nil}, nil
	}

	user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &pb.FindSSLCertUserResponse{
			User: &pb.User{
				Id: userId,
			},
		}, nil
	}

	return &pb.FindSSLCertUserResponse{
		User: &pb.User{
			Id:       userId,
			Username: user.Username,
			Fullname: user.Fullname,
		},
	}, nil
}
