package models

// NodeClusterMetricItem 集群使用的指标
type NodeClusterMetricItem struct {
	Id        uint32 `field:"id"`        // ID
	IsOn      bool   `field:"isOn"`      // 是否启用
	ClusterId uint32 `field:"clusterId"` // 集群ID
	ItemId    uint64 `field:"itemId"`    // 指标ID
	State     uint8  `field:"state"`     // 是否启用
}

type NodeClusterMetricItemOperator struct {
	Id        interface{} // ID
	IsOn      interface{} // 是否启用
	ClusterId interface{} // 集群ID
	ItemId    interface{} // 指标ID
	State     interface{} // 是否启用
}

func NewNodeClusterMetricItemOperator() *NodeClusterMetricItemOperator {
	return &NodeClusterMetricItemOperator{}
}
