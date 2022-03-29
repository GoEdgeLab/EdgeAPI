// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build !plus
// +build !plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// UpdateServerUAM 修改服务UAM设置
func (this *ServerService) UpdateServerUAM(ctx context.Context, req *pb.UpdateServerUAMRequest) (*pb.RPCSuccess, error) {
	return this.Success()
}

// FindEnabledServerUAM 查找服务UAM设置
func (this *ServerService) FindEnabledServerUAM(ctx context.Context, req *pb.FindEnabledServerUAMRequest) (*pb.FindEnabledServerUAMResponse, error) {
	return &pb.FindEnabledServerUAMResponse{
		UamJSON: nil,
	}, nil
}
