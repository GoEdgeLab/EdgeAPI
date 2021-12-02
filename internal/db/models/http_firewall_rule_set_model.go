package models

// HTTPFirewallRuleSet 防火墙规则集
type HTTPFirewallRuleSet struct {
	Id            uint32 `field:"id"`            // ID
	IsOn          uint8  `field:"isOn"`          // 是否启用
	Code          string `field:"code"`          // 代号
	Name          string `field:"name"`          // 名称
	Description   string `field:"description"`   // 描述
	CreatedAt     uint64 `field:"createdAt"`     // 创建时间
	Rules         string `field:"rules"`         // 规则列表
	Connector     string `field:"connector"`     // 规则之间的关系
	State         uint8  `field:"state"`         // 状态
	AdminId       uint32 `field:"adminId"`       // 管理员ID
	UserId        uint32 `field:"userId"`        // 用户ID
	Action        string `field:"action"`        // 执行的动作（过期）
	ActionOptions string `field:"actionOptions"` // 动作的选项（过期）
	Actions       string `field:"actions"`       // 一组动作
	IgnoreLocal   uint8  `field:"ignoreLocal"`   // 忽略局域网请求
}

type HTTPFirewallRuleSetOperator struct {
	Id            interface{} // ID
	IsOn          interface{} // 是否启用
	Code          interface{} // 代号
	Name          interface{} // 名称
	Description   interface{} // 描述
	CreatedAt     interface{} // 创建时间
	Rules         interface{} // 规则列表
	Connector     interface{} // 规则之间的关系
	State         interface{} // 状态
	AdminId       interface{} // 管理员ID
	UserId        interface{} // 用户ID
	Action        interface{} // 执行的动作（过期）
	ActionOptions interface{} // 动作的选项（过期）
	Actions       interface{} // 一组动作
	IgnoreLocal   interface{} // 忽略局域网请求
}

func NewHTTPFirewallRuleSetOperator() *HTTPFirewallRuleSetOperator {
	return &HTTPFirewallRuleSetOperator{}
}
