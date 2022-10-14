package models

import "github.com/iwind/TeaGo/dbs"

// NodeRegion 节点区域
type NodeRegion struct {
	Id          uint32   `field:"id"`          // ID
	AdminId     uint32   `field:"adminId"`     // 管理员ID
	IsOn        bool     `field:"isOn"`        // 是否启用
	Name        string   `field:"name"`        // 名称
	Description string   `field:"description"` // 描述
	Order       uint32   `field:"order"`       // 排序
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
	Prices      dbs.JSON `field:"prices"`      // 流量价格
	State       uint8    `field:"state"`       // 状态
}

type NodeRegionOperator struct {
	Id          any // ID
	AdminId     any // 管理员ID
	IsOn        any // 是否启用
	Name        any // 名称
	Description any // 描述
	Order       any // 排序
	CreatedAt   any // 创建时间
	Prices      any // 流量价格
	State       any // 状态
}

func NewNodeRegionOperator() *NodeRegionOperator {
	return &NodeRegionOperator{}
}
