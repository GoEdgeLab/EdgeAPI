package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ACME任务相关服务
type ACMETaskService struct {
	BaseService
}

// 计算某个ACME用户相关的任务数量
func (this *ACMETaskService) CountAllEnabledACMETasksWithACMEUserId(ctx context.Context, req *pb.CountAllEnabledACMETasksWithACMEUserIdRequest) (*pb.RPCCountResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	count, err := models.SharedACMETaskDAO.CountACMETasksWithACMEUserId(req.AcmeUserId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 计算跟某个DNS服务商相关的任务数量
func (this *ACMETaskService) CountEnabledACMETasksWithDNSProviderId(ctx context.Context, req *pb.CountEnabledACMETasksWithDNSProviderIdRequest) (*pb.RPCCountResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	count, err := models.SharedACMETaskDAO.CountACMETasksWithDNSProviderId(req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 计算所有任务数量
func (this *ACMETaskService) CountAllEnabledACMETasks(ctx context.Context, req *pb.CountAllEnabledACMETasksRequest) (*pb.RPCCountResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedACMETaskDAO.CountAllEnabledACMETasks(req.AdminId, req.UserId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 列出单页任务
func (this *ACMETaskService) ListEnabledACMETasks(ctx context.Context, req *pb.ListEnabledACMETasksRequest) (*pb.ListEnabledACMETasksResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	tasks, err := models.SharedACMETaskDAO.ListEnabledACMETasks(req.AdminId, req.UserId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.ACMETask{}
	for _, task := range tasks {
		// ACME用户
		acmeUser, err := models.SharedACMEUserDAO.FindEnabledACMEUser(int64(task.AcmeUserId))
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

		// DNS
		provider, err := models.SharedDNSProviderDAO.FindEnabledDNSProvider(int64(task.DnsProviderId))
		if err != nil {
			return nil, err
		}
		if provider == nil {
			continue
		}
		pbProvider := &pb.DNSProvider{
			Id:       int64(provider.Id),
			Name:     provider.Name,
			Type:     provider.Type,
			TypeName: dnsclients.FindProviderTypeName(provider.Type),
		}

		// 证书
		var pbCert *pb.SSLCert = nil
		if task.CertId > 0 {
			cert, err := models.SharedSSLCertDAO.FindEnabledSSLCert(int64(task.CertId))
			if err != nil {
				return nil, err
			}
			if cert == nil {
				continue
			}
			pbCert = &pb.SSLCert{
				Id:          int64(cert.Id),
				IsOn:        cert.IsOn == 1,
				Name:        cert.Name,
				TimeBeginAt: int64(cert.TimeBeginAt),
				TimeEndAt:   int64(cert.TimeEndAt),
			}
		}

		// 最近一条日志
		var pbTaskLog *pb.ACMETaskLog = nil
		taskLog, err := models.SharedACMETaskLogDAO.FindLatestACMETasKLog(int64(task.Id))
		if err != nil {
			return nil, err
		}
		if taskLog != nil {
			pbTaskLog = &pb.ACMETaskLog{
				Id:        int64(taskLog.Id),
				IsOk:      taskLog.IsOk == 1,
				Error:     taskLog.Error,
				CreatedAt: int64(taskLog.CreatedAt),
			}
		}

		result = append(result, &pb.ACMETask{
			Id:                int64(task.Id),
			IsOn:              task.IsOn == 1,
			DnsDomain:         task.DnsDomain,
			Domains:           task.DecodeDomains(),
			CreatedAt:         int64(task.CreatedAt),
			AutoRenew:         task.AutoRenew == 1,
			AcmeUser:          pbACMEUser,
			DnsProvider:       pbProvider,
			SslCert:           pbCert,
			LatestACMETaskLog: pbTaskLog,
		})
	}

	return &pb.ListEnabledACMETasksResponse{AcmeTasks: result}, nil
}

// 创建任务
func (this *ACMETaskService) CreateACMETask(ctx context.Context, req *pb.CreateACMETaskRequest) (*pb.CreateACMETaskResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0)
	if err != nil {
		return nil, err
	}
	taskId, err := models.SharedACMETaskDAO.CreateACMETask(adminId, userId, req.AcmeUserId, req.DnsProviderId, req.DnsDomain, req.Domains, req.AutoRenew)
	if err != nil {
		return nil, err
	}
	return &pb.CreateACMETaskResponse{AcmeTaskId: taskId}, nil
}

// 修改任务
func (this *ACMETaskService) UpdateACMETask(ctx context.Context, req *pb.UpdateACMETaskRequest) (*pb.RPCSuccess, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0)
	if err != nil {
		return nil, err
	}

	canAccess, err := models.SharedACMETaskDAO.CheckACMETask(adminId, userId, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, this.PermissionError()
	}

	err = models.SharedACMETaskDAO.UpdateACMETask(req.AcmeTaskId, req.AcmeUserId, req.DnsProviderId, req.DnsDomain, req.Domains, req.AutoRenew)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 删除任务
func (this *ACMETaskService) DeleteACMETask(ctx context.Context, req *pb.DeleteACMETaskRequest) (*pb.RPCSuccess, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0)
	if err != nil {
		return nil, err
	}

	canAccess, err := models.SharedACMETaskDAO.CheckACMETask(adminId, userId, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, this.PermissionError()
	}

	err = models.SharedACMETaskDAO.DisableACMETask(req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 运行某个任务
func (this *ACMETaskService) RunACMETask(ctx context.Context, req *pb.RunACMETaskRequest) (*pb.RunACMETaskResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0)
	if err != nil {
		return nil, err
	}

	canAccess, err := models.SharedACMETaskDAO.CheckACMETask(adminId, userId, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, this.PermissionError()
	}

	isOk, msg, certId := models.SharedACMETaskDAO.RunTask(req.AcmeTaskId)

	return &pb.RunACMETaskResponse{
		IsOk:      isOk,
		Error:     msg,
		SslCertId: certId,
	}, nil
}

// 查找单个任务信息
func (this *ACMETaskService) FindEnabledACMETask(ctx context.Context, req *pb.FindEnabledACMETaskRequest) (*pb.FindEnabledACMETaskResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0)
	if err != nil {
		return nil, err
	}

	canAccess, err := models.SharedACMETaskDAO.CheckACMETask(adminId, userId, req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, this.PermissionError()
	}

	task, err := models.SharedACMETaskDAO.FindEnabledACMETask(req.AcmeTaskId)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return &pb.FindEnabledACMETaskResponse{AcmeTask: nil}, nil
	}

	// 用户
	var pbACMEUser *pb.ACMEUser = nil
	if task.AcmeUserId > 0 {
		acmeUser, err := models.SharedACMEUserDAO.FindEnabledACMEUser(int64(task.AcmeUserId))
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
		}
	}

	// DNS
	var pbProvider *pb.DNSProvider
	provider, err := models.SharedDNSProviderDAO.FindEnabledDNSProvider(int64(task.DnsProviderId))
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

	return &pb.FindEnabledACMETaskResponse{AcmeTask: &pb.ACMETask{
		Id:          int64(task.Id),
		IsOn:        task.IsOn == 1,
		DnsDomain:   task.DnsDomain,
		Domains:     task.DecodeDomains(),
		CreatedAt:   int64(task.CreatedAt),
		AutoRenew:   task.AutoRenew == 1,
		DnsProvider: pbProvider,
		AcmeUser:    pbACMEUser,
	}}, nil
}
