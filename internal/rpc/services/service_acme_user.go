package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 用户服务
type ACMEUserService struct {
	BaseService
}

// 创建用户
func (this *ACMEUserService) CreateACMEUser(ctx context.Context, req *pb.CreateACMEUserRequest) (*pb.CreateACMEUserResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	acmeUserId, err := models.SharedACMEUserDAO.CreateACMEUser(adminId, userId, req.Email, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateACMEUserResponse{AcmeUserId: acmeUserId}, nil
}

// 修改用户
func (this *ACMEUserService) UpdateACMEUser(ctx context.Context, req *pb.UpdateACMEUserRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	// 检查是否有权限
	b, err := models.SharedACMEUserDAO.CheckACMEUser(req.AcmeUserId, adminId, userId)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, this.PermissionError()
	}

	err = models.SharedACMEUserDAO.UpdateACMEUser(req.AcmeUserId, req.Description)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 删除用户
func (this *ACMEUserService) DeleteACMEUser(ctx context.Context, req *pb.DeleteACMEUserRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	// 检查是否有权限
	b, err := models.SharedACMEUserDAO.CheckACMEUser(req.AcmeUserId, adminId, userId)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, this.PermissionError()
	}

	err = models.SharedACMEUserDAO.DisableACMEUser(req.AcmeUserId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 计算用户数量
func (this *ACMEUserService) CountACMEUsers(ctx context.Context, req *pb.CountAcmeUsersRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedACMEUserDAO.CountACMEUsersWithAdminId(adminId, userId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// 列出单页用户
func (this *ACMEUserService) ListACMEUsers(ctx context.Context, req *pb.ListACMEUsersRequest) (*pb.ListACMEUsersResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	acmeUsers, err := models.SharedACMEUserDAO.ListACMEUsers(adminId, userId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.ACMEUser{}
	for _, user := range acmeUsers {
		result = append(result, &pb.ACMEUser{
			Id:          int64(user.Id),
			Email:       user.Email,
			Description: user.Description,
			CreatedAt:   int64(user.CreatedAt),
		})
	}
	return &pb.ListACMEUsersResponse{AcmeUsers: result}, nil
}

// 查找单个用户
func (this *ACMEUserService) FindEnabledACMEUser(ctx context.Context, req *pb.FindEnabledACMEUserRequest) (*pb.FindEnabledACMEUserResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	// 检查是否有权限
	b, err := models.SharedACMEUserDAO.CheckACMEUser(req.AcmeUserId, adminId, userId)
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, this.PermissionError()
	}

	acmeUser, err := models.SharedACMEUserDAO.FindEnabledACMEUser(req.AcmeUserId)
	if err != nil {
		return nil, err
	}
	if acmeUser == nil {
		return &pb.FindEnabledACMEUserResponse{AcmeUser: nil}, nil
	}
	return &pb.FindEnabledACMEUserResponse{AcmeUser: &pb.ACMEUser{
		Id:          int64(acmeUser.Id),
		Email:       acmeUser.Email,
		Description: acmeUser.Description,
		CreatedAt:   int64(acmeUser.CreatedAt),
	}}, nil
}
