package models

// HTTP防火墙
type HTTPFirewallPolicy struct {
	Id          uint32 `field:"id"`          // ID
	TemplateId  uint32 `field:"templateId"`  // 模版ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	UserId      uint32 `field:"userId"`      // 用户ID
	State       uint8  `field:"state"`       // 状态
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	IsOn        uint8  `field:"isOn"`        // 是否启用
	Name        string `field:"name"`        // 名称
	Description string `field:"description"` // 描述
	Inbound     string `field:"inbound"`     // 入站规则
	Outbound    string `field:"outbound"`    // 出站规则
}

type HTTPFirewallPolicyOperator struct {
	Id          interface{} // ID
	TemplateId  interface{} // 模版ID
	AdminId     interface{} // 管理员ID
	UserId      interface{} // 用户ID
	State       interface{} // 状态
	CreatedAt   interface{} // 创建时间
	IsOn        interface{} // 是否启用
	Name        interface{} // 名称
	Description interface{} // 描述
	Inbound     interface{} // 入站规则
	Outbound    interface{} // 出站规则
}

func NewHTTPFirewallPolicyOperator() *HTTPFirewallPolicyOperator {
	return &HTTPFirewallPolicyOperator{}
}
