package models

import "github.com/iwind/TeaGo/dbs"

// NodeRegion 节点区域
type NodeRegion struct {
	Id          uint32   `field:"id"`          // ID
	AdminId     uint32   `field:"adminId"`     // 管理员ID
	IsOn        uint8    `field:"isOn"`        // 是否启用
	Name        string   `field:"name"`        // 名称
	Description string   `field:"description"` // 描述
	Order       uint32   `field:"order"`       // 排序
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
	Prices      dbs.JSON `field:"prices"`      // 价格
	State       uint8    `field:"state"`       // 状态
}

type NodeRegionOperator struct {
	Id          interface{} // ID
	AdminId     interface{} // 管理员ID
	IsOn        interface{} // 是否启用
	Name        interface{} // 名称
	Description interface{} // 描述
	Order       interface{} // 排序
	CreatedAt   interface{} // 创建时间
	Prices      interface{} // 价格
	State       interface{} // 状态
}

func NewNodeRegionOperator() *NodeRegionOperator {
	return &NodeRegionOperator{}
}
