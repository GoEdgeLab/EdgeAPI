package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 用户AccessKey相关服务
type UserAccessKeyService struct {
	BaseService
}

// 创建AccessKey
func (this *UserAccessKeyService) CreateUserAccessKey(ctx context.Context, req *pb.CreateUserAccessKeyRequest) (*pb.CreateUserAccessKeyResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	userAccessKeyId, err := models.SharedUserAccessKeyDAO.CreateAccessKey(req.UserId, req.Description)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserAccessKeyResponse{UserAccessKeyId: userAccessKeyId}, nil
}

// 查找所有的AccessKey
func (this *UserAccessKeyService) FindAllEnabledUserAccessKeys(ctx context.Context, req *pb.FindAllEnabledUserAccessKeysRequest) (*pb.FindAllEnabledUserAccessKeysResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	accessKeys, err := models.SharedUserAccessKeyDAO.FindAllEnabledAccessKeys(req.UserId)
	if err != nil {
		return nil, err
	}

	result := []*pb.UserAccessKey{}
	for _, accessKey := range accessKeys {
		result = append(result, &pb.UserAccessKey{
			Id:          int64(accessKey.Id),
			UserId:      int64(accessKey.UserId),
			SubUserId:   int64(accessKey.SubUserId),
			IsOn:        accessKey.IsOn == 1,
			UniqueId:    accessKey.UniqueId,
			Secret:      accessKey.Secret,
			Description: accessKey.Description,
		})
	}

	return &pb.FindAllEnabledUserAccessKeysResponse{UserAccessKeys: result}, nil
}

// 删除AccessKey
func (this *UserAccessKeyService) DeleteUserAccessKey(ctx context.Context, req *pb.DeleteUserAccessKeyRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		ok, err := models.SharedUserAccessKeyDAO.CheckUserAccessKey(userId, req.UserAccessKeyId)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, this.PermissionError()
		}
	}

	err = models.SharedUserAccessKeyDAO.DisableUserAccessKey(req.UserAccessKeyId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 设置是否启用AccessKey
func (this *UserAccessKeyService) UpdateUserAccessKeyIsOn(ctx context.Context, req *pb.UpdateUserAccessKeyIsOnRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		ok, err := models.SharedUserAccessKeyDAO.CheckUserAccessKey(userId, req.UserAccessKeyId)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, this.PermissionError()
		}
	}

	err = models.SharedUserAccessKeyDAO.UpdateAccessKeyIsOn(req.UserAccessKeyId, req.IsOn)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
