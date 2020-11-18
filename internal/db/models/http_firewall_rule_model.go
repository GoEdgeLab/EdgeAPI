package models

// 防火墙规则
type HTTPFirewallRule struct {
	Id                uint32 `field:"id"`                // ID
	IsOn              uint8  `field:"isOn"`              // 是否启用
	Description       string `field:"description"`       // 说明
	Param             string `field:"param"`             // 参数
	Filters           string `field:"filters"`           // 处理器
	Operator          string `field:"operator"`          // 操作符
	Value             string `field:"value"`             // 对比值
	IsCaseInsensitive uint8  `field:"isCaseInsensitive"` // 是否大小写不敏感
	CheckpointOptions string `field:"checkpointOptions"` // 检查点参数
	State             uint8  `field:"state"`             // 状态
	CreatedAt         uint64 `field:"createdAt"`         // 创建时间
	AdminId           uint32 `field:"adminId"`           // 管理员ID
	UserId            uint32 `field:"userId"`            // 用户ID
}

type HTTPFirewallRuleOperator struct {
	Id                interface{} // ID
	IsOn              interface{} // 是否启用
	Description       interface{} // 说明
	Param             interface{} // 参数
	Filters           interface{} // 处理器
	Operator          interface{} // 操作符
	Value             interface{} // 对比值
	IsCaseInsensitive interface{} // 是否大小写不敏感
	CheckpointOptions interface{} // 检查点参数
	State             interface{} // 状态
	CreatedAt         interface{} // 创建时间
	AdminId           interface{} // 管理员ID
	UserId            interface{} // 用户ID
}

func NewHTTPFirewallRuleOperator() *HTTPFirewallRuleOperator {
	return &HTTPFirewallRuleOperator{}
}
