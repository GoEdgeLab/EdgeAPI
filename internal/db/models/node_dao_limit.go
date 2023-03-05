// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import (
	"github.com/iwind/TeaGo/dbs"
)

func (this *NodeDAO) CountAllAuthorityNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Where("clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)").
		Count()
}

func (this *NodeDAO) CheckNodesLimit(tx *dbs.Tx) error {
	return nil
}
