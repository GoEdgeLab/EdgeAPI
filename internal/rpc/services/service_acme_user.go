package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/acme"
	acmemodels "github.com/TeaOSLab/EdgeAPI/internal/db/models/acme"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ACMEUserService 用户服务
type ACMEUserService struct {
	BaseService
}

// CreateACMEUser 创建用户
func (this *ACMEUserService) CreateACMEUser(ctx context.Context, req *pb.CreateACMEUserRequest) (*pb.CreateACMEUserResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	acmeUserId, err := acmemodels.SharedACMEUserDAO.CreateACMEUser(tx, adminId, userId, req.AcmeProviderCode, req.AcmeProviderAccountId, req.Email, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateACMEUserResponse{AcmeUserId: acmeUserId}, nil
}

// UpdateACMEUser 修改用户
func (this *ACMEUserService) UpdateACMEUser(ctx context.Context, req *pb.UpdateACMEUserRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查是否有权限
	b, err := acmemodels.SharedACMEUserDAO.CheckACMEUser(tx, req.AcmeUserId, adminId, userId)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, this.PermissionError()
	}

	err = acmemodels.SharedACMEUserDAO.UpdateACMEUser(tx, req.AcmeUserId, req.Description)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteACMEUser 删除用户
func (this *ACMEUserService) DeleteACMEUser(ctx context.Context, req *pb.DeleteACMEUserRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查是否有权限
	b, err := acmemodels.SharedACMEUserDAO.CheckACMEUser(tx, req.AcmeUserId, adminId, userId)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, this.PermissionError()
	}

	err = acmemodels.SharedACMEUserDAO.DisableACMEUser(tx, req.AcmeUserId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountACMEUsers 计算用户数量
func (this *ACMEUserService) CountACMEUsers(ctx context.Context, req *pb.CountAcmeUsersRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := acmemodels.SharedACMEUserDAO.CountACMEUsersWithAdminId(tx, adminId, userId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListACMEUsers 列出单页用户
func (this *ACMEUserService) ListACMEUsers(ctx context.Context, req *pb.ListACMEUsersRequest) (*pb.ListACMEUsersResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	acmeUsers, err := acmemodels.SharedACMEUserDAO.ListACMEUsers(tx, adminId, userId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.ACMEUser{}
	for _, user := range acmeUsers {
		var pbUser = &pb.ACMEUser{
			Id:               int64(user.Id),
			Email:            user.Email,
			Description:      user.Description,
			CreatedAt:        int64(user.CreatedAt),
			AcmeProviderCode: user.ProviderCode,
		}

		// 服务商
		if len(user.ProviderCode) == 0 {
			user.ProviderCode = acme.DefaultProviderCode
		}
		var provider = acme.FindProviderWithCode(user.ProviderCode)
		if provider != nil {
			pbUser.AcmeProvider = &pb.ACMEProvider{
				Name:           provider.Name,
				Code:           provider.Code,
				Description:    provider.Description,
				RequireEAB:     provider.RequireEAB,
				EabDescription: provider.EABDescription,
			}
		}

		// 账号
		if user.AccountId > 0 {
			account, err := acmemodels.SharedACMEProviderAccountDAO.FindEnabledACMEProviderAccount(tx, int64(user.AccountId))
			if err != nil {
				return nil, err
			}
			if account != nil {
				pbUser.AcmeProviderAccount = &pb.ACMEProviderAccount{
					Id:           int64(account.Id),
					Name:         account.Name,
					IsOn:         account.IsOn == 1,
					ProviderCode: account.ProviderCode,
					AcmeProvider: nil,
				}

				var provider = acme.FindProviderWithCode(account.ProviderCode)
				if provider != nil {
					pbUser.AcmeProviderAccount.AcmeProvider = &pb.ACMEProvider{
						Name:           provider.Name,
						Code:           provider.Code,
						Description:    provider.Description,
						RequireEAB:     provider.RequireEAB,
						EabDescription: provider.EABDescription,
					}
				}
			}
		}

		result = append(result, pbUser)
	}
	return &pb.ListACMEUsersResponse{AcmeUsers: result}, nil
}

// FindEnabledACMEUser 查找单个用户
func (this *ACMEUserService) FindEnabledACMEUser(ctx context.Context, req *pb.FindEnabledACMEUserRequest) (*pb.FindEnabledACMEUserResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 检查是否有权限
	b, err := acmemodels.SharedACMEUserDAO.CheckACMEUser(tx, req.AcmeUserId, adminId, userId)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, this.PermissionError()
	}

	acmeUser, err := acmemodels.SharedACMEUserDAO.FindEnabledACMEUser(tx, req.AcmeUserId)
	if err != nil {
		return nil, err
	}
	if acmeUser == nil {
		return &pb.FindEnabledACMEUserResponse{AcmeUser: nil}, nil
	}

	// 服务商
	var pbACMEUser = &pb.ACMEUser{
		Id:               int64(acmeUser.Id),
		Email:            acmeUser.Email,
		Description:      acmeUser.Description,
		CreatedAt:        int64(acmeUser.CreatedAt),
		AcmeProviderCode: acmeUser.ProviderCode,
	}
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
				IsOn:         account.IsOn == 1,
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

	return &pb.FindEnabledACMEUserResponse{AcmeUser: pbACMEUser}, nil
}

// FindAllACMEUsers 查找所有用户
func (this *ACMEUserService) FindAllACMEUsers(ctx context.Context, req *pb.FindAllACMEUsersRequest) (*pb.FindAllACMEUsersResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	acmeUsers, err := acmemodels.SharedACMEUserDAO.FindAllACMEUsers(tx, adminId, userId, req.AcmeProviderCode)
	if err != nil {
		return nil, err
	}
	result := []*pb.ACMEUser{}
	for _, user := range acmeUsers {
		result = append(result, &pb.ACMEUser{
			Id:               int64(user.Id),
			Email:            user.Email,
			Description:      user.Description,
			CreatedAt:        int64(user.CreatedAt),
			AcmeProviderCode: user.ProviderCode,
		})
	}
	return &pb.FindAllACMEUsersResponse{AcmeUsers: result}, nil
}
