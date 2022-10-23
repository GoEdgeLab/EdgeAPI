package models

// NodeTask 节点同步任务
type NodeTask struct {
	Id         uint64 `field:"id"`         // ID
	Role       string `field:"role"`       // 节点角色
	NodeId     uint32 `field:"nodeId"`     // 节点ID
	ClusterId  uint32 `field:"clusterId"`  // 集群ID
	ServerId   uint64 `field:"serverId"`   // 服务ID
	UserId     uint64 `field:"userId"`     // 用户ID
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
	Id         any // ID
	Role       any // 节点角色
	NodeId     any // 节点ID
	ClusterId  any // 集群ID
	ServerId   any // 服务ID
	UserId     any // 用户ID
	Type       any // 任务类型
	UniqueId   any // 唯一ID：nodeId@type
	UpdatedAt  any // 修改时间
	IsDone     any // 是否已完成
	IsOk       any // 是否已完成
	Error      any // 错误信息
	IsNotified any // 是否已通知更新
	Version    any // 版本
}

func NewNodeTaskOperator() *NodeTaskOperator {
	return &NodeTaskOperator{}
}
