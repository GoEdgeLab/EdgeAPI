// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

//go:build community
// +build community

package accesslogs

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 写入日志
func (this *StorageManager) Write(policyId int64, accessLogs []*pb.HTTPAccessLog) error {
	return nil
}
