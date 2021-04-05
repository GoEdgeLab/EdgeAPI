package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 互斥锁管理
type SysLockerService struct {
	BaseService
}

// 获得锁
func (this *SysLockerService) SysLockerLock(ctx context.Context, req *pb.SysLockerLockRequest) (*pb.SysLockerLockResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		_, err = this.ValidateMonitor(ctx)
		if err != nil {
			return nil, err
		}
	}

	key := req.Key
	if userId > 0 {
		key = "@user" // 这里不加入用户ID，防止多个用户间冲突
	}

	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 60
	} else if timeout > 86400 { // 最多不能超过1天
		timeout = 86400
	}

	var tx = this.NullTx()
	ok, err := models.SharedSysLockerDAO.Lock(tx, key, timeout)
	if err != nil {
		return nil, err
	}
	return &pb.SysLockerLockResponse{Ok: ok}, nil
}

// 释放锁
func (this *SysLockerService) SysLockerUnlock(ctx context.Context, req *pb.SysLockerUnlockRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		_, err = this.ValidateMonitor(ctx)
		if err != nil {
			return nil, err
		}
	}

	key := req.Key
	if userId > 0 {
		key = "@user"
	}
	var tx = this.NullTx()
	err = models.SharedSysLockerDAO.Unlock(tx, key)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
