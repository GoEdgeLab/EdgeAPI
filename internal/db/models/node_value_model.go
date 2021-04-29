package models

// NodeValue 节点监控数据
type NodeValue struct {
	Id        uint64 `field:"id"`        // ID
	NodeId    uint32 `field:"nodeId"`    // 节点ID
	Role      string `field:"role"`      // 节点角色
	Item      string `field:"item"`      // 监控项
	Value     string `field:"value"`     // 数据
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	Day       string `field:"day"`       // 日期
	Hour      string `field:"hour"`      // 小时
	Minute    string `field:"minute"`    // 分钟
}

type NodeValueOperator struct {
	Id        interface{} // ID
	NodeId    interface{} // 节点ID
	Role      interface{} // 节点角色
	Item      interface{} // 监控项
	Value     interface{} // 数据
	CreatedAt interface{} // 创建时间
	Day       interface{} // 日期
	Hour      interface{} // 小时
	Minute    interface{} // 分钟
}

func NewNodeValueOperator() *NodeValueOperator {
	return &NodeValueOperator{}
}
