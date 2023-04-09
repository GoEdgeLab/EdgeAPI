package models

import "github.com/iwind/TeaGo/dbs"

// HTTPWebsocket Websocket设置
type HTTPWebsocket struct {
	Id                uint32   `field:"id"`                // ID
	AdminId           uint32   `field:"adminId"`           // 管理员ID
	UserId            uint32   `field:"userId"`            // 用户ID
	CreatedAt         uint64   `field:"createdAt"`         // 创建时间
	State             uint8    `field:"state"`             // 状态
	IsOn              bool     `field:"isOn"`              // 是否启用
	HandshakeTimeout  dbs.JSON `field:"handshakeTimeout"`  // 握手超时时间
	AllowAllOrigins   uint8    `field:"allowAllOrigins"`   // 是否支持所有源
	AllowedOrigins    dbs.JSON `field:"allowedOrigins"`    // 支持的源域名列表
	RequestSameOrigin uint8    `field:"requestSameOrigin"` // 是否请求一样的Origin
	RequestOrigin     string   `field:"requestOrigin"`     // 请求Origin
	WebId             uint64   `field:"webId"`             // Web
}

type HTTPWebsocketOperator struct {
	Id                any // ID
	AdminId           any // 管理员ID
	UserId            any // 用户ID
	CreatedAt         any // 创建时间
	State             any // 状态
	IsOn              any // 是否启用
	HandshakeTimeout  any // 握手超时时间
	AllowAllOrigins   any // 是否支持所有源
	AllowedOrigins    any // 支持的源域名列表
	RequestSameOrigin any // 是否请求一样的Origin
	RequestOrigin     any // 请求Origin
	WebId             any // Web
}

func NewHTTPWebsocketOperator() *HTTPWebsocketOperator {
	return &HTTPWebsocketOperator{}
}
