package models

// TCP防火墙
type TCPFirewallPolicy struct {
	Id         uint32 `field:"id"`         // ID
	AdminId    int32  `field:"adminId"`    // 管理员ID
	UserId     uint32 `field:"userId"`     // 用户ID
	TemplateId uint32 `field:"templateId"` // 模版ID
}

type TCPFirewallPolicyOperator struct {
	Id         interface{} // ID
	AdminId    interface{} // 管理员ID
	UserId     interface{} // 用户ID
	TemplateId interface{} // 模版ID
}

func NewTCPFirewallPolicyOperator() *TCPFirewallPolicyOperator {
	return &TCPFirewallPolicyOperator{}
}
