package models

// 访问日志策略
type HTTPAccessLogPolicy struct {
	Id         uint32 `field:"id"`         // ID
	TemplateId uint32 `field:"templateId"` // 模版ID
	AdminId    uint32 `field:"adminId"`    // 管理员ID
	UserId     uint32 `field:"userId"`     // 用户ID
	State      uint8  `field:"state"`      // 状态
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	Name       string `field:"name"`       // 名称
	IsOn       uint8  `field:"isOn"`       // 是否启用
	Type       string `field:"type"`       // 存储类型
	Options    string `field:"options"`    // 存储选项
	Conds      string `field:"conds"`      // 请求条件
}

type HTTPAccessLogPolicyOperator struct {
	Id         interface{} // ID
	TemplateId interface{} // 模版ID
	AdminId    interface{} // 管理员ID
	UserId     interface{} // 用户ID
	State      interface{} // 状态
	CreatedAt  interface{} // 创建时间
	Name       interface{} // 名称
	IsOn       interface{} // 是否启用
	Type       interface{} // 存储类型
	Options    interface{} // 存储选项
	Conds      interface{} // 请求条件
}

func NewHTTPAccessLogPolicyOperator() *HTTPAccessLogPolicyOperator {
	return &HTTPAccessLogPolicyOperator{}
}
