// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build community
// +build community

package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
)

func (this *NodeDAO) composeExtConfig(tx *dbs.Tx, config *nodeconfigs.NodeConfig, cacheMap *utils.CacheMap) error {
	return nil
}
