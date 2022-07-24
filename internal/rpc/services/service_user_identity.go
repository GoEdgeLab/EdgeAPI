// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
)

// UserIdentityService 用户身份认证服务
type UserIdentityService struct {
	BaseService
}

// CreateUserIdentity 创建身份认证信息
func (this *UserIdentityService) CreateUserIdentity(ctx context.Context, req *pb.CreateUserIdentityRequest) (*pb.CreateUserIdentityResponse, error) {
	userId, err := this.ValidateUserNode(ctx)
	if err != nil {
		return nil, err
	}

	switch req.Type {
	case userconfigs.UserIdentityTypeIDCard:
		if len(req.FileIds) < 2 {
			return nil, errors.New("need for file(s) for id card")
		}
	case userconfigs.UserIdentityTypeEnterpriseLicense:
		if len(req.FileIds) != 1 {
			return nil, errors.New("need for file(s) for license")
		}
	default:
		return nil, errors.New("unknown identity type '" + req.Type + "'")
	}

	var tx = this.NullTx()
	identityId, err := models.SharedUserIdentityDAO.CreateUserIdentity(tx, userId, req.OrgType, req.Type, req.RealName, req.Number, req.FileIds)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserIdentityResponse{UserIdentityId: identityId}, nil
}

// FindEnabledUserIdentity 查找单个身份认证信息
func (this *UserIdentityService) FindEnabledUserIdentity(ctx context.Context, req *pb.FindEnabledUserIdentityRequest) (*pb.FindEnabledUserIdentityResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedUserIdentityDAO.CheckUserIdentity(tx, userId, req.UserIdentityId)
		if err != nil {
			return nil, err
		}
	}

	identity, err := models.SharedUserIdentityDAO.FindEnabledUserIdentity(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}
	if identity == nil {
		return &pb.FindEnabledUserIdentityResponse{
			UserIdentity: nil,
		}, nil
	}

	return &pb.FindEnabledUserIdentityResponse{
		UserIdentity: &pb.UserIdentity{
			Id:             int64(identity.Id),
			Type:           identity.Type,
			RealName:       identity.RealName,
			Number:         identity.Number,
			FileIds:        identity.DecodeFileIds(),
			Status:         identity.Status,
			CreatedAt:      int64(identity.CreatedAt),
			UpdatedAt:      int64(identity.UpdatedAt),
			SubmittedAt:    int64(identity.SubmittedAt),
			RejectedAt:     int64(identity.RejectedAt),
			VerifiedAt:     int64(identity.VerifiedAt),
			RejectedReason: identity.RejectedReason,
		},
	}, nil
}

// FindEnabledUserIdentityWithOrgType 查看最新的身份认证信息
func (this *UserIdentityService) FindEnabledUserIdentityWithOrgType(ctx context.Context, req *pb.FindEnabledUserIdentityWithOrgTypeRequest) (*pb.FindEnabledUserIdentityWithOrgTypeResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	identity, err := models.SharedUserIdentityDAO.FindEnabledUserIdentityWithOrgType(tx, req.UserId, req.OrgType)
	if err != nil {
		return nil, err
	}
	if identity == nil {
		return &pb.FindEnabledUserIdentityWithOrgTypeResponse{
			UserIdentity: nil,
		}, nil
	}

	return &pb.FindEnabledUserIdentityWithOrgTypeResponse{
		UserIdentity: &pb.UserIdentity{
			Id:             int64(identity.Id),
			Type:           identity.Type,
			RealName:       identity.RealName,
			Number:         identity.Number,
			FileIds:        identity.DecodeFileIds(),
			Status:         identity.Status,
			CreatedAt:      int64(identity.CreatedAt),
			UpdatedAt:      int64(identity.UpdatedAt),
			SubmittedAt:    int64(identity.SubmittedAt),
			RejectedAt:     int64(identity.RejectedAt),
			VerifiedAt:     int64(identity.VerifiedAt),
			RejectedReason: identity.RejectedReason,
		},
	}, nil
}

// UpdateUserIdentity 修改身份认证信息
func (this *UserIdentityService) UpdateUserIdentity(ctx context.Context, req *pb.UpdateUserIdentityRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if len(req.RealName) > 100 {
		return nil, errors.New("realName too long")
	}

	switch req.Type {
	case userconfigs.UserIdentityTypeIDCard:
		if len(req.FileIds) < 2 {
			return nil, errors.New("need for file(s) for id card")
		}
	case userconfigs.UserIdentityTypeEnterpriseLicense:
		if len(req.FileIds) != 1 {
			return nil, errors.New("need for file(s) for license")
		}
	default:
		return nil, errors.New("unknown identity type '" + req.Type + "'")
	}

	// 检查用户
	err = models.SharedUserIdentityDAO.CheckUserIdentity(tx, userId, req.UserIdentityId)
	if err != nil {
		return nil, err
	}

	// 检查状态
	status, err := models.SharedUserIdentityDAO.FindUserIdentityStatus(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}
	if len(status) > 0 && (status != userconfigs.UserIdentityStatusNone && status != userconfigs.UserIdentityStatusRejected) {
		return nil, errors.New("identity status should be '" + userconfigs.UserIdentityStatusNone + "' instead of '" + status + "'")
	}

	err = models.SharedUserIdentityDAO.UpdateUserIdentity(tx, req.UserIdentityId, req.Type, req.RealName, req.Number, req.FileIds)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// SubmitUserIdentity 提交审核身份认证信息
func (this *UserIdentityService) SubmitUserIdentity(ctx context.Context, req *pb.SubmitUserIdentityRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户
	err = models.SharedUserIdentityDAO.CheckUserIdentity(tx, userId, req.UserIdentityId)
	if err != nil {
		return nil, err
	}

	// 检查状态
	status, err := models.SharedUserIdentityDAO.FindUserIdentityStatus(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}
	if len(status) > 0 && status != userconfigs.UserIdentityStatusNone && status != userconfigs.UserIdentityStatusRejected {
		return nil, errors.New("identity status should be '" + userconfigs.UserIdentityStatusNone + "' instead of '" + status + "'")
	}

	err = models.SharedUserIdentityDAO.SubmitUserIdentity(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// CancelUserIdentity 取消提交身份审核认证信息
func (this *UserIdentityService) CancelUserIdentity(ctx context.Context, req *pb.CancelUserIdentityRequest) (*pb.RPCSuccess, error) {
	userId, err := this.ValidateUserNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户
	err = models.SharedUserIdentityDAO.CheckUserIdentity(tx, userId, req.UserIdentityId)
	if err != nil {
		return nil, err
	}

	// 检查状态
	status, err := models.SharedUserIdentityDAO.FindUserIdentityStatus(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}
	if status != userconfigs.UserIdentityStatusSubmitted {
		return nil, errors.New("identity status should be '" + userconfigs.UserIdentityStatusSubmitted + "' instead of '" + status + "'")
	}

	err = models.SharedUserIdentityDAO.CancelUserIdentity(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ResetUserIdentity 重置提交身份审核认证信息
func (this *UserIdentityService) ResetUserIdentity(ctx context.Context, req *pb.ResetUserIdentityRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedUserIdentityDAO.ResetUserIdentity(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// RejectUserIdentity 拒绝用户身份认证信息
func (this *UserIdentityService) RejectUserIdentity(ctx context.Context, req *pb.RejectUserIdentityRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查状态
	status, err := models.SharedUserIdentityDAO.FindUserIdentityStatus(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}
	if status != userconfigs.UserIdentityStatusSubmitted {
		return nil, errors.New("identity status should be '" + userconfigs.UserIdentityStatusSubmitted + "' instead of '" + status + "'")
	}

	err = models.SharedUserIdentityDAO.RejectUserIdentity(tx, req.UserIdentityId, req.Reason)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// VerifyUserIdentity 通过用户身份认证信息
func (this *UserIdentityService) VerifyUserIdentity(ctx context.Context, req *pb.VerifyUserIdentityRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查状态
	status, err := models.SharedUserIdentityDAO.FindUserIdentityStatus(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}
	if status != userconfigs.UserIdentityStatusSubmitted {
		return nil, errors.New("identity status should be '" + userconfigs.UserIdentityStatusSubmitted + "' instead of '" + status + "'")
	}

	err = models.SharedUserIdentityDAO.VerifyUserIdentity(tx, req.UserIdentityId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
