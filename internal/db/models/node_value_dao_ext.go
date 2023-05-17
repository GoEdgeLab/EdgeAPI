// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
)

// 节点值变更Hook
func (this *NodeValueDAO) nodeValueHook(tx *dbs.Tx, role nodeconfigs.NodeRole, nodeId int64, item nodeconfigs.NodeValueItem, valueJSON []byte) error {
	return nil
}
