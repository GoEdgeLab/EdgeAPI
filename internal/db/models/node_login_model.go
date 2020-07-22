package models

//
type NodeLogin struct {
	Id     uint32 `field:"id"`     // ID
	NodeId uint32 `field:"nodeId"` // 节点ID
	Name   string `field:"name"`   // 名称
	Type   string `field:"type"`   // 类型：ssh,agent
	Params string `field:"params"` // 配置参数
	State  uint8  `field:"state"`  // 状态
}

type NodeLoginOperator struct {
	Id     interface{} // ID
	NodeId interface{} // 节点ID
	Name   interface{} // 名称
	Type   interface{} // 类型：ssh,agent
	Params interface{} // 配置参数
	State  interface{} // 状态
}

func NewNodeLoginOperator() *NodeLoginOperator {
	return &NodeLoginOperator{}
}
