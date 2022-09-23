// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
)

func (this *NodeLoginDAO) FindFrequentGrantIds(tx *dbs.Tx, nodeClusterId int64, nsClusterId int64) ([]int64, error) {
	var query = this.Query(tx).
		Attr("state", NodeLoginStateEnabled).
		Result("JSON_EXTRACT(params, '$.grantId') as `grantId`", "COUNT(*) AS c").
		Having("grantId>0").
		Desc("c").
		Limit(3).
		Group("grantId")
	if nodeClusterId > 0 {
		query.Attr("role", nodeconfigs.NodeRoleNode)
		query.Where("(nodeId IN (SELECT id FROM "+SharedNodeDAO.Table+" WHERE state=1 AND clusterId=:clusterId))").
			Param("clusterId", nodeClusterId)
	} else if nsClusterId > 0 {
		return nil, nil
	}
	ones, _, err := query.
		FindOnes()
	if err != nil {
		return nil, err
	}
	var grantIds = []int64{}
	for _, one := range ones {
		grantIds = append(grantIds, one.GetInt64("grantId"))
	}
	return grantIds, nil
}
