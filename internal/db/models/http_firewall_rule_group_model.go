package models

import "github.com/iwind/TeaGo/dbs"

// HTTPFirewallRuleGroup 防火墙规则分组
type HTTPFirewallRuleGroup struct {
	Id          uint32   `field:"id"`          // ID
	IsOn        bool     `field:"isOn"`        // 是否启用
	Name        string   `field:"name"`        // 名称
	Description string   `field:"description"` // 描述
	Code        string   `field:"code"`        // 代号
	IsTemplate  uint8    `field:"isTemplate"`  // 是否为预置模板
	AdminId     uint32   `field:"adminId"`     // 管理员ID
	UserId      uint32   `field:"userId"`      // 用户ID
	State       uint8    `field:"state"`       // 状态
	Sets        dbs.JSON `field:"sets"`        // 规则集列表
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
}

type HTTPFirewallRuleGroupOperator struct {
	Id          interface{} // ID
	IsOn        interface{} // 是否启用
	Name        interface{} // 名称
	Description interface{} // 描述
	Code        interface{} // 代号
	IsTemplate  interface{} // 是否为预置模板
	AdminId     interface{} // 管理员ID
	UserId      interface{} // 用户ID
	State       interface{} // 状态
	Sets        interface{} // 规则集列表
	CreatedAt   interface{} // 创建时间
}

func NewHTTPFirewallRuleGroupOperator() *HTTPFirewallRuleGroupOperator {
	return &HTTPFirewallRuleGroupOperator{}
}
