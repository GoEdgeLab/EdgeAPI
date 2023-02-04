// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// LoginSessionService 登录SESSION服务
type LoginSessionService struct {
	BaseService
}

// CreateLoginSession 创建SESSION
func (this *LoginSessionService) CreateLoginSession(ctx context.Context, req *pb.CreateLoginSessionRequest) (*pb.RPCSuccess, error) {
	if len(req.Sid) == 0 {
		return nil, errors.New("'sid' should not be empty")
	}

	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	_, err = models.SharedLoginSessionDAO.CreateSession(tx, req.Sid, req.Ip, req.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// WriteLoginSessionValue 写入SESSION数据
func (this *LoginSessionService) WriteLoginSessionValue(ctx context.Context, req *pb.WriteLoginSessionValueRequest) (*pb.RPCSuccess, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedLoginSessionDAO.WriteSessionValue(tx, req.Sid, req.Key, req.Value)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteLoginSession 删除SESSION
func (this *LoginSessionService) DeleteLoginSession(ctx context.Context, req *pb.DeleteLoginSessionRequest) (*pb.RPCSuccess, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if len(req.Sid) == 0 {
		return nil, errors.New("'sid' should not be empty")
	}

	var tx = this.NullTx()
	err = models.SharedLoginSessionDAO.DeleteSession(tx, req.Sid)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindLoginSession 查找SESSION
func (this *LoginSessionService) FindLoginSession(ctx context.Context, req *pb.FindLoginSessionRequest) (*pb.FindLoginSessionResponse, error) {
	_, _, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	if len(req.Sid) == 0 {
		return nil, errors.New("'token' should not be empty")
	}

	var tx = this.NullTx()
	session, err := models.SharedLoginSessionDAO.FindSession(tx, req.Sid)
	if err != nil {
		return nil, err
	}
	if session == nil || !session.IsAvailable() {
		return &pb.FindLoginSessionResponse{
			LoginSession: nil,
		}, nil
	}

	return &pb.FindLoginSessionResponse{
		LoginSession: &pb.LoginSession{
			Id:         int64(session.Id),
			Sid:        session.Sid,
			AdminId:    int64(session.AdminId),
			UserId:     int64(session.UserId),
			Ip:         session.Ip,
			CreatedAt:  int64(session.CreatedAt),
			ExpiresAt:  int64(session.ExpiresAt),
			ValuesJSON: session.Values,
		},
	}, nil
}
