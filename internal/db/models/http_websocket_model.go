package models

// Websocket设置
type HTTPWebsocket struct {
	Id                uint32 `field:"id"`                // ID
	AdminId           uint32 `field:"adminId"`           // 管理员ID
	UserId            uint32 `field:"userId"`            // 用户ID
	CreatedAt         uint64 `field:"createdAt"`         // 创建时间
	State             uint8  `field:"state"`             // 状态
	IsOn              uint8  `field:"isOn"`              // 是否启用
	HandshakeTimeout  string `field:"handshakeTimeout"`  // 握手超时时间
	AllowAllOrigins   uint8  `field:"allowAllOrigins"`   // 是否支持所有源
	AllowedOrigins    string `field:"allowedOrigins"`    // 支持的源域名列表
	RequestSameOrigin uint8  `field:"requestSameOrigin"` // 是否请求一样的Origin
	RequestOrigin     string `field:"requestOrigin"`     // 请求Origin
}

type HTTPWebsocketOperator struct {
	Id                interface{} // ID
	AdminId           interface{} // 管理员ID
	UserId            interface{} // 用户ID
	CreatedAt         interface{} // 创建时间
	State             interface{} // 状态
	IsOn              interface{} // 是否启用
	HandshakeTimeout  interface{} // 握手超时时间
	AllowAllOrigins   interface{} // 是否支持所有源
	AllowedOrigins    interface{} // 支持的源域名列表
	RequestSameOrigin interface{} // 是否请求一样的Origin
	RequestOrigin     interface{} // 请求Origin
}

func NewHTTPWebsocketOperator() *HTTPWebsocketOperator {
	return &HTTPWebsocketOperator{}
}
