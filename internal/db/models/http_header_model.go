package models

//
type HTTPHeader struct {
	Id         uint32 `field:"id"`         // ID
	AdminId    uint32 `field:"adminId"`    // 管理员ID
	UserId     uint32 `field:"userId"`     // 用户ID
	TemplateId uint32 `field:"templateId"` // 模版ID
	IsOn       uint8  `field:"isOn"`       // 是否启用
	Name       string `field:"name"`       // 名称
	Value      string `field:"value"`      // 值
	Order      uint32 `field:"order"`      // 排序
	Status     string `field:"status"`     // 状态码设置
	State      uint8  `field:"state"`      // 状态
	CreatedAt  uint32 `field:"createdAt"`  // 创建时间
}

type HTTPHeaderOperator struct {
	Id         interface{} // ID
	AdminId    interface{} // 管理员ID
	UserId     interface{} // 用户ID
	TemplateId interface{} // 模版ID
	IsOn       interface{} // 是否启用
	Name       interface{} // 名称
	Value      interface{} // 值
	Order      interface{} // 排序
	Status     interface{} // 状态码设置
	State      interface{} // 状态
	CreatedAt  interface{} // 创建时间
}

func NewHTTPHeaderOperator() *HTTPHeaderOperator {
	return &HTTPHeaderOperator{}
}
