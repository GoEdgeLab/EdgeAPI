package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// UserAccessKeyService 用户AccessKey相关服务
type UserAccessKeyService struct {
	BaseService
}

// CreateUserAccessKey 创建AccessKey
func (this *UserAccessKeyService) CreateUserAccessKey(ctx context.Context, req *pb.CreateUserAccessKeyRequest) (*pb.CreateUserAccessKeyResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	userAccessKeyId, err := models.SharedUserAccessKeyDAO.CreateAccessKey(tx, req.AdminId, req.UserId, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserAccessKeyResponse{UserAccessKeyId: userAccessKeyId}, nil
}

// FindAllEnabledUserAccessKeys 查找所有的AccessKey
func (this *UserAccessKeyService) FindAllEnabledUserAccessKeys(ctx context.Context, req *pb.FindAllEnabledUserAccessKeysRequest) (*pb.FindAllEnabledUserAccessKeysResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	accessKeys, err := models.SharedUserAccessKeyDAO.FindAllEnabledAccessKeys(tx, req.AdminId, req.UserId)
	if err != nil {
		return nil, err
	}

	result := []*pb.UserAccessKey{}
	for _, accessKey := range accessKeys {
		result = append(result, &pb.UserAccessKey{
			Id:          int64(accessKey.Id),
			UserId:      int64(accessKey.UserId),
			SubUserId:   int64(accessKey.SubUserId),
			IsOn:        accessKey.IsOn,
			UniqueId:    accessKey.UniqueId,
			Secret:      accessKey.Secret,
			Description: accessKey.Description,
			AccessedAt:  int64(accessKey.AccessedAt),
		})
	}

	return &pb.FindAllEnabledUserAccessKeysResponse{UserAccessKeys: result}, nil
}

// DeleteUserAccessKey 删除AccessKey
func (this *UserAccessKeyService) DeleteUserAccessKey(ctx context.Context, req *pb.DeleteUserAccessKeyRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		ok, err := models.SharedUserAccessKeyDAO.CheckUserAccessKey(tx, 0, userId, req.UserAccessKeyId)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, this.PermissionError()
		}
	}

	err = models.SharedUserAccessKeyDAO.DisableUserAccessKey(tx, req.UserAccessKeyId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateUserAccessKeyIsOn 设置是否启用AccessKey
func (this *UserAccessKeyService) UpdateUserAccessKeyIsOn(ctx context.Context, req *pb.UpdateUserAccessKeyIsOnRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		ok, err := models.SharedUserAccessKeyDAO.CheckUserAccessKey(tx, 0, userId, req.UserAccessKeyId)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, this.PermissionError()
		}
	}

	err = models.SharedUserAccessKeyDAO.UpdateAccessKeyIsOn(tx, req.UserAccessKeyId, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledUserAccessKeys 计算AccessKey数量
func (this *UserAccessKeyService) CountAllEnabledUserAccessKeys(ctx context.Context, req *pb.CountAllEnabledUserAccessKeysRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		req.UserId = userId
	}

	count, err := models.SharedUserAccessKeyDAO.CountAllEnabledAccessKeys(tx, req.AdminId, req.UserId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}
