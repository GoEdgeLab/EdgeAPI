package models

// HTTPHeader HTTP Header
type HTTPHeader struct {
	Id              uint32 `field:"id"`              // ID
	AdminId         uint32 `field:"adminId"`         // 管理员ID
	UserId          uint32 `field:"userId"`          // 用户ID
	TemplateId      uint32 `field:"templateId"`      // 模版ID
	IsOn            uint8  `field:"isOn"`            // 是否启用
	Name            string `field:"name"`            // 名称
	Value           string `field:"value"`           // 值
	Order           uint32 `field:"order"`           // 排序
	Status          string `field:"status"`          // 状态码设置
	DisableRedirect uint8  `field:"disableRedirect"` // 是否不支持跳转
	ShouldAppend    uint8  `field:"shouldAppend"`    // 是否为附加
	ShouldReplace   uint8  `field:"shouldReplace"`   // 是否替换变量
	ReplaceValues   string `field:"replaceValues"`   // 替换的值
	Methods         string `field:"methods"`         // 支持的方法
	Domains         string `field:"domains"`         // 支持的域名
	State           uint8  `field:"state"`           // 状态
	CreatedAt       uint64 `field:"createdAt"`       // 创建时间
}

type HTTPHeaderOperator struct {
	Id              interface{} // ID
	AdminId         interface{} // 管理员ID
	UserId          interface{} // 用户ID
	TemplateId      interface{} // 模版ID
	IsOn            interface{} // 是否启用
	Name            interface{} // 名称
	Value           interface{} // 值
	Order           interface{} // 排序
	Status          interface{} // 状态码设置
	DisableRedirect interface{} // 是否不支持跳转
	ShouldAppend    interface{} // 是否为附加
	ShouldReplace   interface{} // 是否替换变量
	ReplaceValues   interface{} // 替换的值
	Methods         interface{} // 支持的方法
	Domains         interface{} // 支持的域名
	State           interface{} // 状态
	CreatedAt       interface{} // 创建时间
}

func NewHTTPHeaderOperator() *HTTPHeaderOperator {
	return &HTTPHeaderOperator{}
}
