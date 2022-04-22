package models

import "github.com/iwind/TeaGo/dbs"

// HTTPAccessLogPolicy 访问日志策略
type HTTPAccessLogPolicy struct {
	Id           uint32   `field:"id"`           // ID
	TemplateId   uint32   `field:"templateId"`   // 模版ID
	AdminId      uint32   `field:"adminId"`      // 管理员ID
	UserId       uint32   `field:"userId"`       // 用户ID
	State        uint8    `field:"state"`        // 状态
	CreatedAt    uint64   `field:"createdAt"`    // 创建时间
	Name         string   `field:"name"`         // 名称
	IsOn         bool     `field:"isOn"`         // 是否启用
	Type         string   `field:"type"`         // 存储类型
	Options      dbs.JSON `field:"options"`      // 存储选项
	Conds        dbs.JSON `field:"conds"`        // 请求条件
	IsPublic     bool     `field:"isPublic"`     // 是否为公用
	FirewallOnly uint8    `field:"firewallOnly"` // 是否只记录防火墙相关
	Version      uint32   `field:"version"`      // 版本号
}

type HTTPAccessLogPolicyOperator struct {
	Id           interface{} // ID
	TemplateId   interface{} // 模版ID
	AdminId      interface{} // 管理员ID
	UserId       interface{} // 用户ID
	State        interface{} // 状态
	CreatedAt    interface{} // 创建时间
	Name         interface{} // 名称
	IsOn         interface{} // 是否启用
	Type         interface{} // 存储类型
	Options      interface{} // 存储选项
	Conds        interface{} // 请求条件
	IsPublic     interface{} // 是否为公用
	FirewallOnly interface{} // 是否只记录防火墙相关
	Version      interface{} // 版本号
}

func NewHTTPAccessLogPolicyOperator() *HTTPAccessLogPolicyOperator {
	return &HTTPAccessLogPolicyOperator{}
}
