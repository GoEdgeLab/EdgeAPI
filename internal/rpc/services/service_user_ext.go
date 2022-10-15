// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
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
