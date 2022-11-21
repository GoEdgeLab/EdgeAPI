// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// PingService Ping服务
// 用来测试连接是否可用
type PingService struct {
	BaseService
}

// Ping 发起Ping
func (this *PingService) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.PingResponse{}, nil
}
