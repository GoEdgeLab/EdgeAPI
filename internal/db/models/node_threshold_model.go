package models

// NodeThreshold 集群阈值设置
type NodeThreshold struct {
	Id           uint64 `field:"id"`           // ID
	ClusterId    uint32 `field:"clusterId"`    // 集群ID
	NodeId       uint32 `field:"nodeId"`       // 节点ID
	IsOn         uint8  `field:"isOn"`         // 是否启用
	Item         string `field:"item"`         // 监控项
	Param        string `field:"param"`        // 参数
	Operator     string `field:"operator"`     // 操作符
	Value        string `field:"value"`        // 对比值
	Message      string `field:"message"`      // 消息内容
	State        uint8  `field:"state"`        // 状态
	Duration     uint32 `field:"duration"`     // 时间段
	DurationUnit string `field:"durationUnit"` // 时间段单位
	SumMethod    string `field:"sumMethod"`    // 聚合方法
	Order        uint32 `field:"order"`        // 排序
}

type NodeThresholdOperator struct {
	Id           interface{} // ID
	ClusterId    interface{} // 集群ID
	NodeId       interface{} // 节点ID
	IsOn         interface{} // 是否启用
	Item         interface{} // 监控项
	Param        interface{} // 参数
	Operator     interface{} // 操作符
	Value        interface{} // 对比值
	Message      interface{} // 消息内容
	State        interface{} // 状态
	Duration     interface{} // 时间段
	DurationUnit interface{} // 时间段单位
	SumMethod    interface{} // 聚合方法
	Order        interface{} // 排序
}

func NewNodeThresholdOperator() *NodeThresholdOperator {
	return &NodeThresholdOperator{}
}
