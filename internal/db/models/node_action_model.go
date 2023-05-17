package models

import "github.com/iwind/TeaGo/dbs"

// NodeAction 节点智能调度设置
type NodeAction struct {
	Id       uint64   `field:"id"`       // ID
	NodeId   uint64   `field:"nodeId"`   // 节点ID
	Role     string   `field:"role"`     // 角色
	IsOn     bool     `field:"isOn"`     // 是否启用
	Conds    dbs.JSON `field:"conds"`    // 条件
	Action   dbs.JSON `field:"action"`   // 动作
	Duration dbs.JSON `field:"duration"` // 持续时间
	Order    uint32   `field:"order"`    // 排序
	State    uint8    `field:"state"`    // 状态
}

type NodeActionOperator struct {
	Id       any // ID
	NodeId   any // 节点ID
	Role     any // 角色
	IsOn     any // 是否启用
	Conds    any // 条件
	Action   any // 动作
	Duration any // 持续时间
	Order    any // 排序
	State    any // 状态
}

func NewNodeActionOperator() *NodeActionOperator {
	return &NodeActionOperator{}
}
