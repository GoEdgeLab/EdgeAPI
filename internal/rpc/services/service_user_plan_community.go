// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build community
// +build community

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// UserPlanService 用户购买的套餐
type UserPlanService struct {
	BaseService
}

// CreateUserPlan 添加已购套餐
func (this *UserPlanService) BuyUserPlan(ctx context.Context, req *pb.BuyUserPlanRequest) (*pb.BuyUserPlanResponse, error) {
	return &pb.BuyUserPlanResponse{UserPlanId: 0}, nil
}

// RenewUserPlan 续费套餐
func (this *UserPlanService) RenewUserPlan(ctx context.Context, req *pb.RenewUserPlanRequest) (*pb.RPCSuccess, error) {
	return this.Success()
}

// FindEnabledUserPlan 查找单个已购套餐信息
func (this *UserPlanService) FindEnabledUserPlan(ctx context.Context, req *pb.FindEnabledUserPlanRequest) (*pb.FindEnabledUserPlanResponse, error) {
	return &pb.FindEnabledUserPlanResponse{UserPlan: nil}, nil
}

// UpdateUserPlan 修改已购套餐
func (this *UserPlanService) UpdateUserPlan(ctx context.Context, req *pb.UpdateUserPlanRequest) (*pb.RPCSuccess, error) {
	return this.Success()
}

// DeleteUserPlan 删除已购套餐
func (this *UserPlanService) DeleteUserPlan(ctx context.Context, req *pb.DeleteUserPlanRequest) (*pb.RPCSuccess, error) {
	return this.Success()
}

// CountAllEnabledUserPlans 计算已购套餐数
func (this *UserPlanService) CountAllEnabledUserPlans(ctx context.Context, req *pb.CountAllEnabledUserPlansRequest) (*pb.RPCCountResponse, error) {
	return this.SuccessCount(0)
}

// ListEnabledUserPlans 列出单页已购套餐
func (this *UserPlanService) ListEnabledUserPlans(ctx context.Context, req *pb.ListEnabledUserPlansRequest) (*pb.ListEnabledUserPlansResponse, error) {
	return &pb.ListEnabledUserPlansResponse{UserPlans: nil}, nil
}

// FindAllEnabledUserPlansForServer 查找所有服务可用的套餐
func (this *UserPlanService) FindAllEnabledUserPlansForServer(ctx context.Context, req *pb.FindAllEnabledUserPlansForServerRequest) (*pb.FindAllEnabledUserPlansForServerResponse, error) {
	return &pb.FindAllEnabledUserPlansForServerResponse{}, nil
}
