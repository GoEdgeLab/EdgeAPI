package models

import "github.com/iwind/TeaGo/dbs"

// HTTPLocation 路由规则配置
type HTTPLocation struct {
	Id           uint32   `field:"id"`           // ID
	TemplateId   uint32   `field:"templateId"`   // 模版ID
	AdminId      uint32   `field:"adminId"`      // 管理员ID
	UserId       uint32   `field:"userId"`       // 用户ID
	ParentId     uint32   `field:"parentId"`     // 父级ID
	State        uint8    `field:"state"`        // 状态
	CreatedAt    uint64   `field:"createdAt"`    // 创建时间
	Pattern      string   `field:"pattern"`      // 匹配规则
	IsOn         bool     `field:"isOn"`         // 是否启用
	Name         string   `field:"name"`         // 名称
	Description  string   `field:"description"`  // 描述
	WebId        uint32   `field:"webId"`        // Web配置ID
	ReverseProxy dbs.JSON `field:"reverseProxy"` // 反向代理
	UrlPrefix    string   `field:"urlPrefix"`    // URL前缀
	IsBreak      uint8    `field:"isBreak"`      // 是否终止匹配
	Conds        dbs.JSON `field:"conds"`        // 匹配条件
	Domains      dbs.JSON `field:"domains"`      // 专属域名
}

type HTTPLocationOperator struct {
	Id           interface{} // ID
	TemplateId   interface{} // 模版ID
	AdminId      interface{} // 管理员ID
	UserId       interface{} // 用户ID
	ParentId     interface{} // 父级ID
	State        interface{} // 状态
	CreatedAt    interface{} // 创建时间
	Pattern      interface{} // 匹配规则
	IsOn         interface{} // 是否启用
	Name         interface{} // 名称
	Description  interface{} // 描述
	WebId        interface{} // Web配置ID
	ReverseProxy interface{} // 反向代理
	UrlPrefix    interface{} // URL前缀
	IsBreak      interface{} // 是否终止匹配
	Conds        interface{} // 匹配条件
	Domains      interface{} // 专属域名
}

func NewHTTPLocationOperator() *HTTPLocationOperator {
	return &HTTPLocationOperator{}
}
