// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package remotelogs

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
)

type DAOInterface interface {
	CreateLog(tx *dbs.Tx, nodeRole nodeconfigs.NodeRole, nodeId int64, serverId int64, originId int64, level string, tag string, description string, createdAt int64) error
}
