package models

import "github.com/iwind/TeaGo/dbs"

// 重写规则
type HTTPRewriteRule struct {
	Id             uint32   `field:"id"`             // ID
	AdminId        uint32   `field:"adminId"`        // 管理员ID
	UserId         uint32   `field:"userId"`         // 用户ID
	TemplateId     uint32   `field:"templateId"`     // 模版ID
	IsOn           uint8    `field:"isOn"`           // 是否启用
	State          uint8    `field:"state"`          // 状态
	CreatedAt      uint64   `field:"createdAt"`      // 创建时间
	Pattern        string   `field:"pattern"`        // 匹配规则
	Replace        string   `field:"replace"`        // 跳转后的地址
	Mode           string   `field:"mode"`           // 替换模式
	RedirectStatus uint32   `field:"redirectStatus"` // 跳转的状态码
	ProxyHost      string   `field:"proxyHost"`      // 代理的主机名
	IsBreak        uint8    `field:"isBreak"`        // 是否终止解析
	WithQuery      uint8    `field:"withQuery"`      // 是否保留URI参数
	Conds          dbs.JSON `field:"conds"`          // 匹配条件
}

type HTTPRewriteRuleOperator struct {
	Id             interface{} // ID
	AdminId        interface{} // 管理员ID
	UserId         interface{} // 用户ID
	TemplateId     interface{} // 模版ID
	IsOn           interface{} // 是否启用
	State          interface{} // 状态
	CreatedAt      interface{} // 创建时间
	Pattern        interface{} // 匹配规则
	Replace        interface{} // 跳转后的地址
	Mode           interface{} // 替换模式
	RedirectStatus interface{} // 跳转的状态码
	ProxyHost      interface{} // 代理的主机名
	IsBreak        interface{} // 是否终止解析
	WithQuery      interface{} // 是否保留URI参数
	Conds          interface{} // 匹配条件
}

func NewHTTPRewriteRuleOperator() *HTTPRewriteRuleOperator {
	return &HTTPRewriteRuleOperator{}
}
