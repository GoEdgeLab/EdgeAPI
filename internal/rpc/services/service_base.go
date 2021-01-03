package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
)

type BaseService struct {
}

// 校验管理员
func (this *BaseService) ValidateAdmin(ctx context.Context, reqAdminId int64) (adminId int64, err error) {
	_, reqUserId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return
	}
	if reqAdminId > 0 && reqUserId != reqAdminId {
		return 0, this.PermissionError()
	}
	return reqUserId, nil
}

// 校验管理员和用户
func (this *BaseService) ValidateAdminAndUser(ctx context.Context, requireAdminId int64, requireUserId int64) (adminId int64, userId int64, err error) {
	reqUserType, reqUserId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return
	}

	adminId = int64(0)
	userId = int64(0)
	switch reqUserType {
	case rpcutils.UserTypeAdmin:
		adminId = reqUserId
		if adminId < 0 { // 允许AdminId = 0
			err = errors.New("invalid 'adminId'")
			return
		}
		if requireAdminId > 0 && adminId != requireAdminId {
			err = this.PermissionError()
			return
		}
	case rpcutils.UserTypeUser:
		userId = reqUserId
		if userId <= 0 {
			err = errors.New("invalid 'userId'")
			return
		}
		if requireUserId > 0 && userId != requireUserId {
			err = this.PermissionError()
			return
		}
	default:
		err = errors.New("invalid user type")
	}

	return
}

// 校验边缘节点
func (this *BaseService) ValidateNode(ctx context.Context) (nodeId int64, err error) {
	_, nodeId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	return
}

// 校验用户节点
func (this *BaseService) ValidateUser(ctx context.Context) (userId int64, err error) {
	_, userId, err = rpcutils.ValidateRequest(ctx, rpcutils.UserTypeUser)
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

// 空的数据库事务
func (this *BaseService) NullTx() *dbs.Tx {
	return nil
}
