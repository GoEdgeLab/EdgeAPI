package models

import "github.com/iwind/TeaGo/dbs"

// HTTPPage 特殊页面
type HTTPPage struct {
	Id         uint32   `field:"id"`         // ID
	AdminId    uint32   `field:"adminId"`    // 管理员ID
	UserId     uint32   `field:"userId"`     // 用户ID
	IsOn       bool     `field:"isOn"`       // 是否启用
	StatusList dbs.JSON `field:"statusList"` // 状态列表
	Url        string   `field:"url"`        // 页面URL
	NewStatus  int32    `field:"newStatus"`  // 新状态码
	State      uint8    `field:"state"`      // 状态
	CreatedAt  uint64   `field:"createdAt"`  // 创建时间
	Body       string   `field:"body"`       // 页面内容
	BodyType   string   `field:"bodyType"`   // 内容类型
}

type HTTPPageOperator struct {
	Id         interface{} // ID
	AdminId    interface{} // 管理员ID
	UserId     interface{} // 用户ID
	IsOn       interface{} // 是否启用
	StatusList interface{} // 状态列表
	Url        interface{} // 页面URL
	NewStatus  interface{} // 新状态码
	State      interface{} // 状态
	CreatedAt  interface{} // 创建时间
	Body       interface{} // 页面内容
	BodyType   interface{} // 内容类型
}

func NewHTTPPageOperator() *HTTPPageOperator {
	return &HTTPPageOperator{}
}
