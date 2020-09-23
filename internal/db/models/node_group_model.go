package models

// 节点分组
type NodeGroup struct {
	Id        uint32 `field:"id"`        // ID
	Name      string `field:"name"`      // 名称
	Order     uint32 `field:"order"`     // 排序
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	State     uint8  `field:"state"`     // 状态
}

type NodeGroupOperator struct {
	Id        interface{} // ID
	Name      interface{} // 名称
	Order     interface{} // 排序
	CreatedAt interface{} // 创建时间
	State     interface{} // 状态
}

func NewNodeGroupOperator() *NodeGroupOperator {
	return &NodeGroupOperator{}
}
