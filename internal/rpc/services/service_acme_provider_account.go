// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	acmeutils "github.com/TeaOSLab/EdgeAPI/internal/acme"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/acme"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ACMEProviderAccountService ACME服务商账号服务
type ACMEProviderAccountService struct {
	BaseService
}

// CreateACMEProviderAccount 创建服务商账号
func (this *ACMEProviderAccountService) CreateACMEProviderAccount(ctx context.Context, req *pb.CreateACMEProviderAccountRequest) (*pb.CreateACMEProviderAccountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	accountId, err := acme.SharedACMEProviderAccountDAO.CreateAccount(tx, userId, req.Name, req.ProviderCode, req.EabKid, req.EabKey)
	if err != nil {
		return nil, err
	}
	return &pb.CreateACMEProviderAccountResponse{
		AcmeProviderAccountId: accountId,
	}, nil
}

// FindAllACMEProviderAccountsWithProviderCode 使用代号查找服务商账号
func (this *ACMEProviderAccountService) FindAllACMEProviderAccountsWithProviderCode(ctx context.Context, req *pb.FindAllACMEProviderAccountsWithProviderCodeRequest) (*pb.FindAllACMEProviderAccountsWithProviderCodeResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	accounts, err := acme.SharedACMEProviderAccountDAO.FindAllEnabledAccountsWithProviderCode(tx, userId, req.AcmeProviderCode)
	if err != nil {
		return nil, err
	}
	var pbAccounts = []*pb.ACMEProviderAccount{}
	for _, account := range accounts {
		var pbProvider *pb.ACMEProvider
		provider := acmeutils.FindProviderWithCode(account.ProviderCode)
		if provider != nil {
			pbProvider = &pb.ACMEProvider{
				Name:        provider.Name,
				Code:        provider.Code,
				Description: provider.Description,
				ApiURL:      provider.APIURL,
				RequireEAB:  provider.RequireEAB,
			}
		}

		pbAccounts = append(pbAccounts, &pb.ACMEProviderAccount{
			Id:           int64(account.Id),
			Name:         account.Name,
			ProviderCode: account.ProviderCode,
			IsOn:         account.IsOn,
			EabKid:       account.EabKid,
			EabKey:       account.EabKey,
			Error:        account.Error,
			AcmeProvider: pbProvider,
		})
	}

	return &pb.FindAllACMEProviderAccountsWithProviderCodeResponse{
		AcmeProviderAccounts: pbAccounts,
	}, nil
}

// UpdateACMEProviderAccount 修改服务商账号
func (this *ACMEProviderAccountService) UpdateACMEProviderAccount(ctx context.Context, req *pb.UpdateACMEProviderAccountRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = acme.SharedACMEProviderAccountDAO.CheckUserAccount(tx, userId, req.AcmeProviderAccountId)
		if err != nil {
			return nil, err
		}
	}

	err = acme.SharedACMEProviderAccountDAO.UpdateAccount(tx, req.AcmeProviderAccountId, req.Name, req.EabKid, req.EabKey)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteACMEProviderAccount 删除服务商账号
func (this *ACMEProviderAccountService) DeleteACMEProviderAccount(ctx context.Context, req *pb.DeleteACMEProviderAccountRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = acme.SharedACMEProviderAccountDAO.CheckUserAccount(tx, userId, req.AcmeProviderAccountId)
		if err != nil {
			return nil, err
		}
	}

	err = acme.SharedACMEProviderAccountDAO.DisableACMEProviderAccount(tx, req.AcmeProviderAccountId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledACMEProviderAccount 查找单个服务商账号
func (this *ACMEProviderAccountService) FindEnabledACMEProviderAccount(ctx context.Context, req *pb.FindEnabledACMEProviderAccountRequest) (*pb.FindEnabledACMEProviderAccountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查权限
	if userId > 0 {
		err = acme.SharedACMEProviderAccountDAO.CheckUserAccount(tx, userId, req.AcmeProviderAccountId)
		if err != nil {
			return nil, err
		}
	}

	account, err := acme.SharedACMEProviderAccountDAO.FindEnabledACMEProviderAccount(tx, req.AcmeProviderAccountId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return &pb.FindEnabledACMEProviderAccountResponse{AcmeProviderAccount: nil}, nil
	}

	var pbProvider *pb.ACMEProvider
	provider := acmeutils.FindProviderWithCode(account.ProviderCode)
	if provider != nil {
		pbProvider = &pb.ACMEProvider{
			Name:           provider.Name,
			Code:           provider.Code,
			Description:    provider.Description,
			ApiURL:         provider.APIURL,
			RequireEAB:     provider.RequireEAB,
			EabDescription: provider.EABDescription,
		}
	}

	return &pb.FindEnabledACMEProviderAccountResponse{AcmeProviderAccount: &pb.ACMEProviderAccount{
		Id:           int64(account.Id),
		Name:         account.Name,
		ProviderCode: account.ProviderCode,
		IsOn:         account.IsOn,
		EabKid:       account.EabKid,
		EabKey:       account.EabKey,
		Error:        account.Error,
		AcmeProvider: pbProvider,
	}}, nil
}

// CountAllEnabledACMEProviderAccounts 计算所有服务商账号数量
func (this *ACMEProviderAccountService) CountAllEnabledACMEProviderAccounts(ctx context.Context, req *pb.CountAllEnabledACMEProviderAccountsRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := acme.SharedACMEProviderAccountDAO.CountAllEnabledAccounts(tx, userId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledACMEProviderAccounts 列出单页服务商账号
func (this *ACMEProviderAccountService) ListEnabledACMEProviderAccounts(ctx context.Context, req *pb.ListEnabledACMEProviderAccountsRequest) (*pb.ListEnabledACMEProviderAccountsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	accounts, err := acme.SharedACMEProviderAccountDAO.ListEnabledAccounts(tx, userId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbAccounts = []*pb.ACMEProviderAccount{}
	for _, account := range accounts {
		var pbProvider *pb.ACMEProvider
		provider := acmeutils.FindProviderWithCode(account.ProviderCode)
		if provider != nil {
			pbProvider = &pb.ACMEProvider{
				Name:           provider.Name,
				Code:           provider.Code,
				Description:    provider.Description,
				ApiURL:         provider.APIURL,
				RequireEAB:     provider.RequireEAB,
				EabDescription: provider.EABDescription,
			}
		}

		pbAccounts = append(pbAccounts, &pb.ACMEProviderAccount{
			Id:           int64(account.Id),
			Name:         account.Name,
			ProviderCode: account.ProviderCode,
			IsOn:         account.IsOn,
			EabKid:       account.EabKid,
			EabKey:       account.EabKey,
			Error:        account.Error,
			AcmeProvider: pbProvider,
		})
	}

	return &pb.ListEnabledACMEProviderAccountsResponse{AcmeProviderAccounts: pbAccounts}, nil
}
