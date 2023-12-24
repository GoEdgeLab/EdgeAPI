// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import (
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

func (this *NodeDAO) CountAllAuthorityNodes(tx *dbs.Tx) (int64, error) {
	return this.Query(tx).
		State(NodeStateEnabled).
		Where("clusterId IN (SELECT id FROM " + SharedNodeClusterDAO.Table + " WHERE state=1)").
		Count()
}

func (this *NodeDAO) CheckNodesLimit(tx *dbs.Tx) error {
	var maxNodes = teaconst.DefaultMaxNodes

	// 检查节点数量
	if maxNodes > 0 {
		count, err := this.CountAllAuthorityNodes(tx)
		if err != nil {
			return err
		}
		if count >= int64(maxNodes) {
			return errors.New("超出最大节点数限制：" + types.String(maxNodes) + "，当前已用：" + types.String(count) + "，请自行修改源码修改此限制（EdgeAPI/internal/const/const_community.go） 或者 购买商业版本授权。")
		}
	}

	return nil
}
