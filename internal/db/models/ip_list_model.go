package models

import "github.com/iwind/TeaGo/dbs"

// IPList IP名单
type IPList struct {
	Id          uint32   `field:"id"`          // ID
	IsOn        bool     `field:"isOn"`        // 是否启用
	Type        string   `field:"type"`        // 类型
	AdminId     uint32   `field:"adminId"`     // 用户ID
	UserId      uint32   `field:"userId"`      // 用户ID
	Name        string   `field:"name"`        // 列表名
	Code        string   `field:"code"`        // 代号
	State       uint8    `field:"state"`       // 状态
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
	Timeout     dbs.JSON `field:"timeout"`     // 默认超时时间
	Actions     dbs.JSON `field:"actions"`     // IP触发的动作
	Description string   `field:"description"` // 描述
	IsPublic    uint8    `field:"isPublic"`    // 是否公用
	IsGlobal    uint8    `field:"isGlobal"`    // 是否全局
}

type IPListOperator struct {
	Id          interface{} // ID
	IsOn        interface{} // 是否启用
	Type        interface{} // 类型
	AdminId     interface{} // 用户ID
	UserId      interface{} // 用户ID
	Name        interface{} // 列表名
	Code        interface{} // 代号
	State       interface{} // 状态
	CreatedAt   interface{} // 创建时间
	Timeout     interface{} // 默认超时时间
	Actions     interface{} // IP触发的动作
	Description interface{} // 描述
	IsPublic    interface{} // 是否公用
	IsGlobal    interface{} // 是否全局
}

func NewIPListOperator() *IPListOperator {
	return &IPListOperator{}
}
