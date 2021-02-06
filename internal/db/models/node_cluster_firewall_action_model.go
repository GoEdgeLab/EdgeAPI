package models

// 防火墙动作
type NodeClusterFirewallAction struct {
	Id         uint32 `field:"id"`         // ID
	AdminId    uint32 `field:"adminId"`    // 管理员ID
	ClusterId  uint32 `field:"clusterId"`  // 集群ID
	Name       string `field:"name"`       // 名称
	EventLevel string `field:"eventLevel"` // 级别
	Type       string `field:"type"`       // 动作类型
	Params     string `field:"params"`     // 参数
	State      uint8  `field:"state"`      // 状态
}

type NodeClusterFirewallActionOperator struct {
	Id         interface{} // ID
	AdminId    interface{} // 管理员ID
	ClusterId  interface{} // 集群ID
	Name       interface{} // 名称
	EventLevel interface{} // 级别
	Type       interface{} // 动作类型
	Params     interface{} // 参数
	State      interface{} // 状态
}

func NewNodeClusterFirewallActionOperator() *NodeClusterFirewallActionOperator {
	return &NodeClusterFirewallActionOperator{}
}
