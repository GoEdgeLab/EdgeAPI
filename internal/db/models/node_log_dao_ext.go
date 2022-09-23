// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
)

func (this *NodeLogDAO) deleteNodeLogsWithCluster(tx *dbs.Tx, role nodeconfigs.NodeRole, clusterId int64) error  {
	return nil
}