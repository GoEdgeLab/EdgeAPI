package models

import "github.com/iwind/TeaGo/dbs"

// HTTPHeaderPolicy Header定义
type HTTPHeaderPolicy struct {
	Id             uint32   `field:"id"`             // ID
	IsOn           bool     `field:"isOn"`           // 是否启用
	State          uint8    `field:"state"`          // 状态
	AdminId        uint32   `field:"adminId"`        // 管理员ID
	UserId         uint32   `field:"userId"`         // 用户ID
	CreatedAt      uint64   `field:"createdAt"`      // 创建时间
	AddHeaders     dbs.JSON `field:"addHeaders"`     // 添加的Header
	AddTrailers    dbs.JSON `field:"addTrailers"`    // 添加的Trailers
	SetHeaders     dbs.JSON `field:"setHeaders"`     // 设置Header
	ReplaceHeaders dbs.JSON `field:"replaceHeaders"` // 替换Header内容
	Expires        dbs.JSON `field:"expires"`        // Expires单独设置
	DeleteHeaders  dbs.JSON `field:"deleteHeaders"`  // 删除的Headers
	Cors           dbs.JSON `field:"cors"`           // CORS配置
}

type HTTPHeaderPolicyOperator struct {
	Id             any // ID
	IsOn           any // 是否启用
	State          any // 状态
	AdminId        any // 管理员ID
	UserId         any // 用户ID
	CreatedAt      any // 创建时间
	AddHeaders     any // 添加的Header
	AddTrailers    any // 添加的Trailers
	SetHeaders     any // 设置Header
	ReplaceHeaders any // 替换Header内容
	Expires        any // Expires单独设置
	DeleteHeaders  any // 删除的Headers
	Cors           any // CORS配置
}

func NewHTTPHeaderPolicyOperator() *HTTPHeaderPolicyOperator {
	return &HTTPHeaderPolicyOperator{}
}
