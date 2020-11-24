package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

type BaseService struct {
}

// 校验管理员和用户
func (this *BaseService) ValidateAdminAndUser(ctx context.Context) (adminId int64, userId int64, err error) {
	reqUserType, reqUserId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return
	}

	adminId = int64(0)
	userId = int64(0)
	switch reqUserType {
	case rpcutils.UserTypeAdmin:
		adminId = reqUserId
	case rpcutils.UserTypeUser:
		userId = reqUserId
	}
	return
}

// 返回成功
func (this *BaseService) Success() (*pb.RPCSuccess, error) {
	return rpcutils.Success()
}

// 返回数字
func (this *BaseService) ResponseCount(count int64) (*pb.RPCCountResponse, error) {
	return &pb.RPCCountResponse{Count: count}, nil
}

// 返回权限错误
func (this *BaseService) PermissionError() error {
	return errors.New("Permission Denied")
}
