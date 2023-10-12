// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// SendMessageTask 发送某个消息任务
func (this *MessageTaskService) SendMessageTask(ctx context.Context, req *pb.SendMessageTaskRequest) (*pb.SendMessageTaskResponse, error) {
	return nil, this.NotImplementedYet()
}
