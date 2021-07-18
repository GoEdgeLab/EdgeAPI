// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package models

// MessageTaskTarget 消息接收对象
// 每个字段不一定都有值
type MessageTaskTarget struct {
	ClusterId int64 // 集群ID
	NodeId    int64 // 节点ID
	ServerId  int64 // 服务ID
}
