package models

import "github.com/iwind/TeaGo/dbs"

// NodeThreshold 集群阈值设置
type NodeThreshold struct {
	Id             uint64   `field:"id"`             // ID
	Role           string   `field:"role"`           // 节点角色
	ClusterId      uint32   `field:"clusterId"`      // 集群ID
	NodeId         uint32   `field:"nodeId"`         // 节点ID
	IsOn           uint8    `field:"isOn"`           // 是否启用
	Item           string   `field:"item"`           // 监控项
	Param          string   `field:"param"`          // 参数
	Operator       string   `field:"operator"`       // 操作符
	Value          dbs.JSON `field:"value"`          // 对比值
	Message        string   `field:"message"`        // 消息内容
	NotifyDuration uint32   `field:"notifyDuration"` // 通知间隔
	NotifiedAt     uint32   `field:"notifiedAt"`     // 上次通知时间
	Duration       uint32   `field:"duration"`       // 时间段
	DurationUnit   string   `field:"durationUnit"`   // 时间段单位
	SumMethod      string   `field:"sumMethod"`      // 聚合方法
	Order          uint32   `field:"order"`          // 排序
	State          uint8    `field:"state"`          // 状态
}

type NodeThresholdOperator struct {
	Id             interface{} // ID
	Role           interface{} // 节点角色
	ClusterId      interface{} // 集群ID
	NodeId         interface{} // 节点ID
	IsOn           interface{} // 是否启用
	Item           interface{} // 监控项
	Param          interface{} // 参数
	Operator       interface{} // 操作符
	Value          interface{} // 对比值
	Message        interface{} // 消息内容
	NotifyDuration interface{} // 通知间隔
	NotifiedAt     interface{} // 上次通知时间
	Duration       interface{} // 时间段
	DurationUnit   interface{} // 时间段单位
	SumMethod      interface{} // 聚合方法
	Order          interface{} // 排序
	State          interface{} // 状态
}

func NewNodeThresholdOperator() *NodeThresholdOperator {
	return &NodeThresholdOperator{}
}
