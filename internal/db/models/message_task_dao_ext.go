// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/dbs"
)

// CreateMessageTasks 从集群、节点或者服务中创建任务
func (this *MessageTaskDAO) CreateMessageTasks(tx *dbs.Tx, role nodeconfigs.NodeRole, clusterId int64, nodeId int64, serverId int64, messageType MessageType, subject string, body string) error {
	return nil
}
