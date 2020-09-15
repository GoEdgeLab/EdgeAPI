package models

//
type HTTPWeb struct {
	Id         uint32 `field:"id"`         // ID
	IsOn       uint8  `field:"isOn"`       // 是否启用
	TemplateId uint32 `field:"templateId"` // 模版ID
	AdminId    uint32 `field:"adminId"`    // 管理员ID
	UserId     uint32 `field:"userId"`     // 用户ID
	State      uint8  `field:"state"`      // 状态
	CreatedAt  uint32 `field:"createdAt"`  // 创建时间
	Root       string `field:"root"`       // 资源根目录
}

type HTTPWebOperator struct {
	Id         interface{} // ID
	IsOn       interface{} // 是否启用
	TemplateId interface{} // 模版ID
	AdminId    interface{} // 管理员ID
	UserId     interface{} // 用户ID
	State      interface{} // 状态
	CreatedAt  interface{} // 创建时间
	Root       interface{} // 资源根目录
}

func NewHTTPWebOperator() *HTTPWebOperator {
	return &HTTPWebOperator{}
}
