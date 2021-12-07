package models

// NodeIPAddressGroup IP地址分组
type NodeIPAddressGroup struct {
	Id    uint32 `field:"id"`    // ID
	Name  string `field:"name"`  // 分组名
	Value string `field:"value"` // IP值
}

type NodeIPAddressGroupOperator struct {
	Id    interface{} // ID
	Name  interface{} // 分组名
	Value interface{} // IP值
}

func NewNodeIPAddressGroupOperator() *NodeIPAddressGroupOperator {
	return &NodeIPAddressGroupOperator{}
}
