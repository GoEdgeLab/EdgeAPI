package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	acmemodels "github.com/TeaOSLab/EdgeAPI/internal/db/models/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ACMETaskService ACME任务相关服务
type ACMETaskService struct {
	BaseService
}

// CountAllEnabledACMETasksWithACMEUserId 计算某个ACME用户相关的任务数量
func (this *ACMETaskService) CountAllEnabledACMETasksWithACMEUserId(ctx context.Context, req *pb.CountAllEnabledACMETasksWithACMEUserIdRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		// TODO 校验权限
	}

	count, err := acmemodels.SharedACMETaskDAO.CountACMETasksWithACMEUserId(tx, req.AcmeUserId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountEnabledACMETasksWithDNSProviderId 计算跟某个DNS服务商相关的任务数量
func (this *ACMETaskService) CountEnabledACMETasksWithDNSProviderId(ctx context.Context, req *pb.CountEnabledACMETasksWithDNSProviderIdRequest) (*pb.RPCCountResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	var tx = this.NullTx()

	count, err := acmemodels.SharedACMETaskDAO.CountACMETasksWithDNSProviderId(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountAllEnabledACMETasks 计算所有任务数量
func (this *ACMETaskService) CountAllEnabledACMETasks(ctx context.Context, req *pb.CountAllEnabledACMETasksRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	count, err := acmemodels.SharedACMETaskDAO.CountAllEnabledACMETasks(tx, req.UserId, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, req.UserOnly)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledACMETasks 列出单页任务
func (this *ACMETaskService) ListEnabledACMETasks(ctx context.Context, req *pb.ListEnabledACMETasksRequest) (*pb.ListEnabledACMETasksResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId
	}

	tasks, err := acmemodels.SharedACMETaskDAO.ListEnabledACMETasks(tx, req.UserId, req.IsAvailable, req.IsExpired, int64(req.ExpiringDays), req.Keyword, req.UserOnly, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var result = []*pb.ACMETask{}
	for _, task := range tasks {
		// ACME用户
		acmeUser, err := acmemodels.SharedACMEUserDAO.FindEnabledACMEUser(tx, int64(task.AcmeUserId))
		if err != nil {
			return nil, err
		}
		if acmeUser == nil {
			continue
		}
		pbACMEUser := &pb.ACMEUser{
			Id:          int64(acmeUser.Id),
			Email:       acmeUser.Email,
			Description: acmeUser.Description,
			CreatedAt:   int64(acmeUser.CreatedAt),
		}

		// 服务商
		if len(acmeUser.ProviderCode) == 0 {
			acmeUser.ProviderCode = acme.DefaultProviderCode
		}
		var provider = acme.FindProviderWithCode(acmeUser.ProviderCode)
		if provider != nil {
			pbACMEUser.AcmeProvider = &pb.ACMEProvider{
				Name:           provider.Name,
				Code:           provider.Code,
				Description:    provider.Description,
				RequireEAB:     provider.RequireEAB,
				EabDescription: provider.EABDescription,
			}
		}

		// 账号
		if acmeUser.AccountId > 0 {
			account, err := acmemodels.SharedACMEProviderAccountDAO.FindEnabledACMEProviderAccount(tx, int64(acmeUser.AccountId))
			if err != nil {
				return nil, err
			}
			if account != nil {
				pbACMEUser.AcmeProviderAccount = &pb.ACMEProviderAccount{
					Id:           int64(account.Id),
					Name:         account.Name,
					IsOn:         account.IsOn,
					ProviderCode: account.ProviderCode,
					AcmeProvider: nil,
				}

				var provider = acme.FindProviderWithCode(account.ProviderCode)
				if provider != nil {
					pbACMEUser.AcmeProviderAccount.AcmeProvider = &pb.ACMEProvider{
						Name:           provider.Name,
						Code:           provider.Code,
						Description:    provider.Description,
						RequireEAB:     provider.RequireEAB,
						EabDescription: provider.EABDescription,
					}
				}
			}
		}

		var pbDNSProvider *pb.DNSProvider
		if task.AuthType == acme.AuthTypeDNS {
			// DNS
			provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(task.DnsProviderId))
			if err != nil {
				return nil, err
			}
			if provider == nil {
				continue
			}
			pbDNSProvider = &pb.DNSProvider{
				Id:       int64(provider.Id),
				Name:     provider.Name,
				Type:     provider.Type,
				TypeName: dnsclients.FindProviderTypeName(provider.Type),
			}
		}

		// 证书
		var pbCert *pb.SSLCert = nil
		if task.CertId > 0 {
			cert, err := models.SharedSSLCertDAO.FindEnabledSSLCert(tx, int64(task.CertId))
			if err != nil {
				return nil, err
			}
			if cert == nil {
				continue
			}
			pbCert = &pb.SSLCert{
				Id:          int64(cert.Id),
				IsOn:        cert.IsOn,
				Name:        cert.Name,
				TimeBeginAt: int64(cert.TimeBeginAt),
				TimeEndAt:   int64(cert.TimeEndAt),
			}
		}

		// 最近一条日志
		var pbTaskLog *pb.ACMETaskLog = nil
		taskLog, err := acmemodels.SharedACMETaskLogDAO.FindLatestACMETasKLog(tx, int64(task.Id))
		if err != nil {
			return nil, err
		}
		if taskLog != nil {
			pbTaskLog = &pb.ACMETaskLog{
				Id:        int64(taskLog.Id),
				IsOk:      taskLog.IsOk,
				Error:     taskLog.Error,
				CreatedAt: int64(taskLog.CreatedAt),
			}
		}

		result = append(result, &pb.ACMETask{
			Id:                int64(task.Id),
			IsOn:              task.IsOn,
			DnsDomain:         task.DnsDomain,
			Domains:           task.DecodeDomains(),
			CreatedAt:         int64(task.CreatedAt),
			AutoRenew:         task.AutoRenew == 1,
			AcmeUser:          pbACMEUser,
			DnsProvider:       pbDNSProvider,
			SslCert:           pbCert,
			LatestACMETaskLog: pbTaskLog,
			AuthType:          task.AuthType,
			AuthURL:           task.AuthURL,
		})
	}

	return &pb.ListEnabledACMETasksResponse{AcmeTasks: result}, nil
}

// CreateACMETask 创建任务
func (this *ACMETaskService) CreateACMETask(ctx context.Context, req *pb.CreateACMETaskRequest) (*pb.CreateACMETaskResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if len(req.AuthType) == 0 {
		req.AuthType = acme.AuthTypeDNS
	}

	if adminId > 0 {
		if req.UserId > 0 {
			userId = req.UserId
		}
	}

	var tx = this.NullTx()
	taskId, err := acmemodels.SharedACMETaskDAO.CreateACMETask(tx, adminId, userId, req.AuthType, req.AcmeUserId, req.DnsProviderId, req.DnsDomain, req.Domains, req.AutoRenew, req.AuthURL)
	if err != nil {
		return nil, err
	}
	return &pb.CreateACMETaskResponse{AcmeTaskId: taskId}, nil
}

// UpdateACMETask 修改任务
func (this *ACMETaskService) UpdateACMETask(ctx context.Context, req *pb.UpdateACMETaskRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	canAccess, err := acmemodels.SharedACMETaskDAO.CheckUserACMETask(tx, userId, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, this.PermissionError()
	}

	err = acmemodels.SharedACMETaskDAO.UpdateACMETask(tx, req.AcmeTaskId, req.AcmeUserId, req.DnsProviderId, req.DnsDomain, req.Domains, req.AutoRenew, req.AuthURL)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteACMETask 删除任务
func (this *ACMETaskService) DeleteACMETask(ctx context.Context, req *pb.DeleteACMETaskRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	canAccess, err := acmemodels.SharedACMETaskDAO.CheckUserACMETask(tx, userId, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, this.PermissionError()
	}

	err = acmemodels.SharedACMETaskDAO.DisableACMETask(tx, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// RunACMETask 运行某个任务
func (this *ACMETaskService) RunACMETask(ctx context.Context, req *pb.RunACMETaskRequest) (*pb.RunACMETaskResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	canAccess, err := acmemodels.SharedACMETaskDAO.CheckUserACMETask(tx, userId, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, this.PermissionError()
	}

	isOk, msg, certId := acmemodels.SharedACMETaskDAO.RunTask(tx, req.AcmeTaskId)

	return &pb.RunACMETaskResponse{
		IsOk:      isOk,
		Error:     msg,
		SslCertId: certId,
	}, nil
}

// FindEnabledACMETask 查找单个任务信息
func (this *ACMETaskService) FindEnabledACMETask(ctx context.Context, req *pb.FindEnabledACMETaskRequest) (*pb.FindEnabledACMETaskResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	canAccess, err := acmemodels.SharedACMETaskDAO.CheckUserACMETask(tx, userId, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, this.PermissionError()
	}

	task, err := acmemodels.SharedACMETaskDAO.FindEnabledACMETask(tx, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return &pb.FindEnabledACMETaskResponse{AcmeTask: nil}, nil
	}

	// 用户
	var pbACMEUser *pb.ACMEUser = nil
	if task.AcmeUserId > 0 {
		acmeUser, err := acmemodels.SharedACMEUserDAO.FindEnabledACMEUser(tx, int64(task.AcmeUserId))
		if err != nil {
			return nil, err
		}
		if acmeUser != nil {
			pbACMEUser = &pb.ACMEUser{
				Id:          int64(acmeUser.Id),
				Email:       acmeUser.Email,
				Description: acmeUser.Description,
				CreatedAt:   int64(acmeUser.CreatedAt),
			}

			// 服务商
			if len(acmeUser.ProviderCode) == 0 {
				acmeUser.ProviderCode = acme.DefaultProviderCode
			}
			var provider = acme.FindProviderWithCode(acmeUser.ProviderCode)
			if provider != nil {
				pbACMEUser.AcmeProvider = &pb.ACMEProvider{
					Name:           provider.Name,
					Code:           provider.Code,
					Description:    provider.Description,
					RequireEAB:     provider.RequireEAB,
					EabDescription: provider.EABDescription,
				}
			}

			// 账号
			if acmeUser.AccountId > 0 {
				account, err := acmemodels.SharedACMEProviderAccountDAO.FindEnabledACMEProviderAccount(tx, int64(acmeUser.AccountId))
				if err != nil {
					return nil, err
				}
				if account != nil {
					pbACMEUser.AcmeProviderAccount = &pb.ACMEProviderAccount{
						Id:           int64(account.Id),
						Name:         account.Name,
						IsOn:         account.IsOn,
						ProviderCode: account.ProviderCode,
						AcmeProvider: nil,
					}

					var provider = acme.FindProviderWithCode(account.ProviderCode)
					if provider != nil {
						pbACMEUser.AcmeProviderAccount.AcmeProvider = &pb.ACMEProvider{
							Name:           provider.Name,
							Code:           provider.Code,
							Description:    provider.Description,
							RequireEAB:     provider.RequireEAB,
							EabDescription: provider.EABDescription,
						}
					}
				}
			}
		}
	}

	// DNS
	var pbProvider *pb.DNSProvider
	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, int64(task.DnsProviderId))
	if err != nil {
		return nil, err
	}
	if provider != nil {
		pbProvider = &pb.DNSProvider{
			Id:       int64(provider.Id),
			Name:     provider.Name,
			Type:     provider.Type,
			TypeName: dnsclients.FindProviderTypeName(provider.Type),
		}
	}

	// 证书
	var pbCert *pb.SSLCert
	if task.CertId > 0 {
		pbCert = &pb.SSLCert{Id: int64(task.CertId)}
	}

	return &pb.FindEnabledACMETaskResponse{AcmeTask: &pb.ACMETask{
		Id:          int64(task.Id),
		IsOn:        task.IsOn,
		DnsDomain:   task.DnsDomain,
		Domains:     task.DecodeDomains(),
		CreatedAt:   int64(task.CreatedAt),
		AutoRenew:   task.AutoRenew == 1,
		DnsProvider: pbProvider,
		AcmeUser:    pbACMEUser,
		AuthType:    task.AuthType,
		AuthURL:     task.AuthURL,
		SslCert:     pbCert,
	}}, nil
}

// FindACMETaskUser 查找任务所属用户
func (this *ACMETaskService) FindACMETaskUser(ctx context.Context, req *pb.FindACMETaskUserRequest) (*pb.FindACMETaskUserResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	userId, err := acmemodels.SharedACMETaskDAO.FindACMETaskUserId(tx, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if userId <= 0 {
		return &pb.FindACMETaskUserResponse{User: nil}, nil
	}

	user, err := models.SharedUserDAO.FindEnabledBasicUser(tx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &pb.FindACMETaskUserResponse{
			User: &pb.User{
				Id: userId,
			},
		}, nil
	}

	return &pb.FindACMETaskUserResponse{
		User: &pb.User{
			Id:       userId,
			Username: user.Username,
			Fullname: user.Fullname,
		},
	}, nil
}
