// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build !plus
// +build !plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// PlanService 套餐相关服务
type PlanService struct {
	BaseService
}

// CreatePlan 创建套餐
func (this *PlanService) CreatePlan(ctx context.Context, req *pb.CreatePlanRequest) (*pb.CreatePlanResponse, error) {
	return &pb.CreatePlanResponse{}, nil
}

// UpdatePlan 修改套餐
func (this *PlanService) UpdatePlan(ctx context.Context, req *pb.UpdatePlanRequest) (*pb.RPCSuccess, error) {
	return this.Success()
}

// DeletePlan 删除套餐
func (this *PlanService) DeletePlan(ctx context.Context, req *pb.DeletePlanRequest) (*pb.RPCSuccess, error) {
	return this.Success()
}

// FindEnabledPlan 查找单个套餐
func (this *PlanService) FindEnabledPlan(ctx context.Context, req *pb.FindEnabledPlanRequest) (*pb.FindEnabledPlanResponse, error) {
	return &pb.FindEnabledPlanResponse{Plan: nil}, nil
}

// CountAllEnabledPlans 计算套餐数量
func (this *PlanService) CountAllEnabledPlans(ctx context.Context, req *pb.CountAllEnabledPlansRequest) (*pb.RPCCountResponse, error) {
	return this.SuccessCount(0)
}

// ListEnabledPlans 列出单页套餐
func (this *PlanService) ListEnabledPlans(ctx context.Context, req *pb.ListEnabledPlansRequest) (*pb.ListEnabledPlansResponse, error) {
	return &pb.ListEnabledPlansResponse{Plans: nil}, nil
}

// SortPlans 对套餐进行排序
func (this *PlanService) SortPlans(ctx context.Context, req *pb.SortPlansRequest) (*pb.RPCSuccess, error) {
	return this.Success()
}

// FindAllAvailablePlans 列出所有可用的套餐
func (this *PlanService) FindAllAvailablePlans(ctx context.Context, req *pb.FindAllAvailablePlansRequest) (*pb.FindAllAvailablePlansResponse, error) {
	return nil, this.NotImplementedYet()
}
