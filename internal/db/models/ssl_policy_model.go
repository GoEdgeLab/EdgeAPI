package models

//
type SSLPolicy struct {
	Id             uint32 `field:"id"`             // ID
	AdminId        uint32 `field:"adminId"`        // 管理员ID
	UserId         uint32 `field:"userId"`         // 用户ID
	IsOn           uint8  `field:"isOn"`           // 是否启用
	Certs          string `field:"certs"`          // 证书列表
	ClientAuthType uint32 `field:"clientAuthType"` // 客户端认证类型
	MinVersion     string `field:"minVersion"`     // 支持的SSL最小版本
	CipherSuites   string `field:"cipherSuites"`   // 加密算法套件
	Hsts           string `field:"hsts"`           // HSTS设置
	Http2Enabled   uint8  `field:"http2Enabled"`   // 是否启用HTTP/2
	State          uint8  `field:"state"`          // 状态
	CreatedAt      uint64 `field:"createdAt"`      // 创建时间
}

type SSLPolicyOperator struct {
	Id             interface{} // ID
	AdminId        interface{} // 管理员ID
	UserId         interface{} // 用户ID
	IsOn           interface{} // 是否启用
	Certs          interface{} // 证书列表
	ClientAuthType interface{} // 客户端认证类型
	MinVersion     interface{} // 支持的SSL最小版本
	CipherSuites   interface{} // 加密算法套件
	Hsts           interface{} // HSTS设置
	Http2Enabled   interface{} // 是否启用HTTP/2
	State          interface{} // 状态
	CreatedAt      interface{} // 创建时间
}

func NewSSLPolicyOperator() *SSLPolicyOperator {
	return &SSLPolicyOperator{}
}
