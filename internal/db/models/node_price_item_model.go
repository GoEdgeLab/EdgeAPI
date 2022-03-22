package models

// NodePriceItem 区域计费设置
type NodePriceItem struct {
	Id        uint32 `field:"id"`        // ID
	IsOn      bool  `field:"isOn"`      // 是否启用
	Type      string `field:"type"`      // 类型：峰值|流量
	Name      string `field:"name"`      // 名称
	BitsFrom  uint64 `field:"bitsFrom"`  // 起始值
	BitsTo    uint64 `field:"bitsTo"`    // 结束值
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	State     uint8  `field:"state"`     // 状态
}

type NodePriceItemOperator struct {
	Id        interface{} // ID
	IsOn      interface{} // 是否启用
	Type      interface{} // 类型：峰值|流量
	Name      interface{} // 名称
	BitsFrom  interface{} // 起始值
	BitsTo    interface{} // 结束值
	CreatedAt interface{} // 创建时间
	State     interface{} // 状态
}

func NewNodePriceItemOperator() *NodePriceItemOperator {
	return &NodePriceItemOperator{}
}
