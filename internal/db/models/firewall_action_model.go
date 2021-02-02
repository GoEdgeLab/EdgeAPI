package models

// 防火墙动作
type FirewallAction struct {
	Id      uint32 `field:"id"`      // ID
	AdminId uint32 `field:"adminId"` // 管理员ID
	Name    string `field:"name"`    // 名称
	Code    string `field:"code"`    // 快速查询代号
	Type    string `field:"type"`    // 动作类型
	Params  string `field:"params"`  // 参数
	State   uint8  `field:"state"`   // 状态
}

type FirewallActionOperator struct {
	Id      interface{} // ID
	AdminId interface{} // 管理员ID
	Name    interface{} // 名称
	Code    interface{} // 快速查询代号
	Type    interface{} // 动作类型
	Params  interface{} // 参数
	State   interface{} // 状态
}

func NewFirewallActionOperator() *FirewallActionOperator {
	return &FirewallActionOperator{}
}
