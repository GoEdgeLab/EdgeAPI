// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
)

// 增加日统计Hook
func (this *NodeTrafficDailyStatDAO) increaseDailyStatHook(tx *dbs.Tx, role nodeconfigs.NodeRole, nodeId int64) error {
	return nil
}
