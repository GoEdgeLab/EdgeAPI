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
func (this *BaseService) ValidateAdminAndUser(ctx context.Context, reqUserId int64) (adminId int64, userId int64, err error) {
	reqUserType, reqUserId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return
	}

	adminId = int64(0)
	userId = int64(0)
	switch reqUserType {
	case rpcutils.UserTypeAdmin:
		adminId = reqUserId
		if adminId <= 0 {
			err = errors.New("invalid 'adminId'")
			return
		}
	case rpcutils.UserTypeUser:
		userId = reqUserId
		if userId <= 0 {
			err = errors.New("invalid 'userId'")
			return
		}

		// 校验权限
		if reqUserId > 0 && reqUserId != userId {
			err = this.PermissionError()
			return
		}
	default:
		err = errors.New("invalid user type")
	}

	return
}

// 返回成功
func (this *BaseService) Success() (*pb.RPCSuccess, error) {
	return &pb.RPCSuccess{}, nil
}

// 返回数字
func (this *BaseService) SuccessCount(count int64) (*pb.RPCCountResponse, error) {
	return &pb.RPCCountResponse{Count: count}, nil
}

// 返回权限错误
func (this *BaseService) PermissionError() error {
	return errors.New("Permission Denied")
}
