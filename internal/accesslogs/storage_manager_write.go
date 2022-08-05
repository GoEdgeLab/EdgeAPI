// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

//go:build !plus

package accesslogs

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 写入日志
func (this *StorageManager) Write(policyId int64, accessLogs []*pb.HTTPAccessLog) (success bool, failMessage string, err error) {
	return false, "only works in plus version", nil
}
