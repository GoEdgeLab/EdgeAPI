// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package users

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
)

// FindUserPriceInfo 读取用户计费信息
func (this *UserService) FindUserPriceInfo(ctx context.Context, req *pb.FindUserPriceInfoRequest) (*pb.FindUserPriceInfoResponse, error) {
	return nil, this.NotImplementedYet()
}

// UpdateUserPriceType 修改用户计费方式
func (this *UserService) UpdateUserPriceType(ctx context.Context, req *pb.UpdateUserPriceTypeRequest) (*pb.RPCSuccess, error) {
	return nil, this.NotImplementedYet()
}

// UpdateUserPricePeriod 修改用户计费周期
func (this *UserService) UpdateUserPricePeriod(ctx context.Context, req *pb.UpdateUserPricePeriodRequest) (*pb.RPCSuccess, error) {
	return nil, this.NotImplementedYet()
}

// RegisterUser 注册用户
func (this *UserService) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	currentUserId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	if currentUserId > 0 {
		return nil, this.PermissionError()
	}

	var tx = this.NullTx()

	// 检查邮箱是否已被使用
	if len(req.Email) > 0 {
		emailUserId, err := models.SharedUserDAO.FindUserIdWithVerifiedEmail(tx, req.Email)
		if err != nil {
			return nil, err
		}
		if emailUserId > 0 {
			return nil, errors.New("the email address '" + req.Email + "' is using by other user")
		}
	}

	// 注册配置
	registerConfig, err := models.SharedSysSettingDAO.ReadUserRegisterConfig(tx)
	if err != nil {
		return nil, err
	}
	if registerConfig == nil || !registerConfig.IsOn {
		return nil, errors.New("the registration has been disabled")
	}

	var requireEmailVerification = false
	var createdUserId int64
	err = this.RunTx(func(tx *dbs.Tx) error {
		// 检查用户名
		exists, err := models.SharedUserDAO.ExistUser(tx, 0, req.Username)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("the username exists already")
		}

		// 创建用户
		userId, err := models.SharedUserDAO.CreateUser(tx, req.Username, req.Password, req.Fullname, req.Mobile, "", req.Email, "", req.Source, registerConfig.ClusterId, registerConfig.Features, req.Ip, !registerConfig.RequireVerification)
		if err != nil {
			return err
		}
		createdUserId = userId

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &pb.RegisterUserResponse{
		UserId:                   createdUserId,
		RequireEmailVerification: requireEmailVerification,
	}, nil
}

// FindUserFeatures 获取用户所有的功能列表
func (this *UserService) FindUserFeatures(ctx context.Context, req *pb.FindUserFeaturesRequest) (*pb.FindUserFeaturesResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}
	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()

	features, err := models.SharedUserDAO.FindUserFeatures(tx, req.UserId)
	if err != nil {
		return nil, err
	}

	result := []*pb.UserFeature{}
	for _, feature := range features {
		result = append(result, feature.ToPB())
	}

	return &pb.FindUserFeaturesResponse{Features: result}, nil
}
