// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus
// +build !plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// UpdateHTTPWebUAM 修改UAM设置
func (this *HTTPWebService) UpdateHTTPWebUAM(ctx context.Context, req *pb.UpdateHTTPWebUAMRequest) (*pb.RPCSuccess, error) {
	return this.Success()
}

// FindHTTPWebUAM 查找UAM设置
func (this *HTTPWebService) FindHTTPWebUAM(ctx context.Context, req *pb.FindHTTPWebUAMRequest) (*pb.FindHTTPWebUAMResponse, error) {
	return &pb.FindHTTPWebUAMResponse{UamJSON: nil}, nil
}
