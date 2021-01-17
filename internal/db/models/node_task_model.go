package models

// 节点同步任务
type NodeTask struct {
	Id         uint64 `field:"id"`         // ID
	NodeId     uint32 `field:"nodeId"`     // 节点ID
	ClusterId  uint32 `field:"clusterId"`  // 集群ID
	Type       string `field:"type"`       // 任务类型
	UniqueId   string `field:"uniqueId"`   // 唯一ID：nodeId@type
	UpdatedAt  uint64 `field:"updatedAt"`  // 修改时间
	IsDone     uint8  `field:"isDone"`     // 是否已完成
	IsOk       uint8  `field:"isOk"`       // 是否已完成
	Error      string `field:"error"`      // 错误信息
	IsNotified uint8  `field:"isNotified"` // 是否已通知更新
}

type NodeTaskOperator struct {
	Id         interface{} // ID
	NodeId     interface{} // 节点ID
	ClusterId  interface{} // 集群ID
	Type       interface{} // 任务类型
	UniqueId   interface{} // 唯一ID：nodeId@type
	UpdatedAt  interface{} // 修改时间
	IsDone     interface{} // 是否已完成
	IsOk       interface{} // 是否已完成
	Error      interface{} // 错误信息
	IsNotified interface{} // 是否已通知更新
}

func NewNodeTaskOperator() *NodeTaskOperator {
	return &NodeTaskOperator{}
}
