package models

// NodeTask 节点同步任务
type NodeTask struct {
	Id         uint64 `field:"id"`         // ID
	Role       string `field:"role"`       // 节点角色
	NodeId     uint32 `field:"nodeId"`     // 节点ID
	ClusterId  uint32 `field:"clusterId"`  // 集群ID
	ServerId   uint32 `field:"serverId"`   // 服务ID
	Type       string `field:"type"`       // 任务类型
	UniqueId   string `field:"uniqueId"`   // 唯一ID：nodeId@type
	UpdatedAt  uint64 `field:"updatedAt"`  // 修改时间
	IsDone     bool   `field:"isDone"`     // 是否已完成
	IsOk       bool   `field:"isOk"`       // 是否已完成
	Error      string `field:"error"`      // 错误信息
	IsNotified bool   `field:"isNotified"` // 是否已通知更新
	Version    uint64 `field:"version"`    // 版本
}

type NodeTaskOperator struct {
	Id         interface{} // ID
	Role       interface{} // 节点角色
	NodeId     interface{} // 节点ID
	ClusterId  interface{} // 集群ID
	ServerId   interface{} // 服务ID
	Type       interface{} // 任务类型
	UniqueId   interface{} // 唯一ID：nodeId@type
	UpdatedAt  interface{} // 修改时间
	IsDone     interface{} // 是否已完成
	IsOk       interface{} // 是否已完成
	Error      interface{} // 错误信息
	IsNotified interface{} // 是否已通知更新
	Version    interface{} // 版本
}

func NewNodeTaskOperator() *NodeTaskOperator {
	return &NodeTaskOperator{}
}
