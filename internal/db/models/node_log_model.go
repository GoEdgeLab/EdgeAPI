package models

// NodeLog 节点日志
type NodeLog struct {
	Id          uint64 `field:"id"`          // ID
	Role        string `field:"role"`        // 节点角色
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	Tag         string `field:"tag"`         // 标签
	Description string `field:"description"` // 描述
	Level       string `field:"level"`       // 级别
	NodeId      uint32 `field:"nodeId"`      // 节点ID
	Day         string `field:"day"`         // 日期
	ServerId    uint32 `field:"serverId"`    // 服务ID
	Hash        string `field:"hash"`        // 信息内容Hash
	Count       uint32 `field:"count"`       // 重复次数
}

type NodeLogOperator struct {
	Id          interface{} // ID
	Role        interface{} // 节点角色
	CreatedAt   interface{} // 创建时间
	Tag         interface{} // 标签
	Description interface{} // 描述
	Level       interface{} // 级别
	NodeId      interface{} // 节点ID
	Day         interface{} // 日期
	ServerId    interface{} // 服务ID
	Hash        interface{} // 信息内容Hash
	Count       interface{} // 重复次数
}

func NewNodeLogOperator() *NodeLogOperator {
	return &NodeLogOperator{}
}
