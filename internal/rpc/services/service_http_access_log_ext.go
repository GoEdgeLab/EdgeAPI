// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package services

import "github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"

func (this *HTTPAccessLogService) writeAccessLogsToPolicy(pbAccessLogs []*pb.HTTPAccessLog) error {
	return nil
}
