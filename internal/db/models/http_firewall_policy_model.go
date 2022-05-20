package models

import "github.com/iwind/TeaGo/dbs"

// HTTPFirewallPolicy HTTP防火墙
type HTTPFirewallPolicy struct {
	Id               uint32   `field:"id"`               // ID
	TemplateId       uint32   `field:"templateId"`       // 模版ID
	AdminId          uint32   `field:"adminId"`          // 管理员ID
	UserId           uint32   `field:"userId"`           // 用户ID
	ServerId         uint32   `field:"serverId"`         // 服务ID
	GroupId          uint32   `field:"groupId"`          // 服务分组ID
	State            uint8    `field:"state"`            // 状态
	CreatedAt        uint64   `field:"createdAt"`        // 创建时间
	IsOn             bool     `field:"isOn"`             // 是否启用
	Name             string   `field:"name"`             // 名称
	Description      string   `field:"description"`      // 描述
	Inbound          dbs.JSON `field:"inbound"`          // 入站规则
	Outbound         dbs.JSON `field:"outbound"`         // 出站规则
	BlockOptions     dbs.JSON `field:"blockOptions"`     // BLOCK选项
	CaptchaOptions   dbs.JSON `field:"captchaOptions"`   // 验证码选项
	Mode             string   `field:"mode"`             // 模式
	UseLocalFirewall uint8    `field:"useLocalFirewall"` // 是否自动使用本地防火墙
	SynFlood         dbs.JSON `field:"synFlood"`         // SynFlood防御设置
	Log              dbs.JSON `field:"log"`              // 日志配置
}

type HTTPFirewallPolicyOperator struct {
	Id               interface{} // ID
	TemplateId       interface{} // 模版ID
	AdminId          interface{} // 管理员ID
	UserId           interface{} // 用户ID
	ServerId         interface{} // 服务ID
	GroupId          interface{} // 服务分组ID
	State            interface{} // 状态
	CreatedAt        interface{} // 创建时间
	IsOn             interface{} // 是否启用
	Name             interface{} // 名称
	Description      interface{} // 描述
	Inbound          interface{} // 入站规则
	Outbound         interface{} // 出站规则
	BlockOptions     interface{} // BLOCK选项
	CaptchaOptions   interface{} // 验证码选项
	Mode             interface{} // 模式
	UseLocalFirewall interface{} // 是否自动使用本地防火墙
	SynFlood         interface{} // SynFlood防御设置
	Log              interface{} // 日志配置
}

func NewHTTPFirewallPolicyOperator() *HTTPFirewallPolicyOperator {
	return &HTTPFirewallPolicyOperator{}
}
