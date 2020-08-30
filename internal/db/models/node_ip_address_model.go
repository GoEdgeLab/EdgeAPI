package models

//
type NodeIPAddress struct {
	Id          uint32 `field:"id"`          // ID
	NodeId      uint32 `field:"nodeId"`      // 节点ID
	Name        string `field:"name"`        // 名称
	IP          string `field:"ip"`          // IP地址
	Description string `field:"description"` // 描述
	State       uint8  `field:"state"`       // 状态
	Order       uint32 `field:"order"`       // 排序
}

type NodeIPAddressOperator struct {
	Id          interface{} // ID
	NodeId      interface{} // 节点ID
	Name        interface{} // 名称
	IP          interface{} // IP地址
	Description interface{} // 描述
	State       interface{} // 状态
	Order       interface{} // 排序
}

func NewNodeIPAddressOperator() *NodeIPAddressOperator {
	return &NodeIPAddressOperator{}
}
