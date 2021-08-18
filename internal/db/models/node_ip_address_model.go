package models

// NodeIPAddress 节点IP地址
type NodeIPAddress struct {
	Id          uint32 `field:"id"`          // ID
	NodeId      uint32 `field:"nodeId"`      // 节点ID
	Role        string `field:"role"`        // 节点角色
	Name        string `field:"name"`        // 名称
	Ip          string `field:"ip"`          // IP地址
	Description string `field:"description"` // 描述
	State       uint8  `field:"state"`       // 状态
	Order       uint32 `field:"order"`       // 排序
	CanAccess   uint8  `field:"canAccess"`   // 是否可以访问
	IsOn        uint8  `field:"isOn"`        // 是否启用
	IsUp        uint8  `field:"isUp"`        // 是否上线
	Thresholds  string `field:"thresholds"`  // 上线阈值
}

type NodeIPAddressOperator struct {
	Id          interface{} // ID
	NodeId      interface{} // 节点ID
	Role        interface{} // 节点角色
	Name        interface{} // 名称
	Ip          interface{} // IP地址
	Description interface{} // 描述
	State       interface{} // 状态
	Order       interface{} // 排序
	CanAccess   interface{} // 是否可以访问
	IsOn        interface{} // 是否启用
	IsUp        interface{} // 是否上线
	Thresholds  interface{} // 上线阈值
}

func NewNodeIPAddressOperator() *NodeIPAddressOperator {
	return &NodeIPAddressOperator{}
}
