package models

import "github.com/iwind/TeaGo/dbs"

// SSLPolicy SSL配置策略
type SSLPolicy struct {
	Id               uint32   `field:"id"`               // ID
	AdminId          uint32   `field:"adminId"`          // 管理员ID
	UserId           uint32   `field:"userId"`           // 用户ID
	IsOn             bool     `field:"isOn"`             // 是否启用
	Certs            dbs.JSON `field:"certs"`            // 证书列表
	ClientCACerts    dbs.JSON `field:"clientCACerts"`    // 客户端证书
	ClientAuthType   uint32   `field:"clientAuthType"`   // 客户端认证类型
	MinVersion       string   `field:"minVersion"`       // 支持的SSL最小版本
	CipherSuitesIsOn uint8    `field:"cipherSuitesIsOn"` // 是否自定义加密算法套件
	CipherSuites     dbs.JSON `field:"cipherSuites"`     // 加密算法套件
	Hsts             dbs.JSON `field:"hsts"`             // HSTS设置
	Http2Enabled     bool     `field:"http2Enabled"`     // 是否启用HTTP/2
	Http3Enabled     bool     `field:"http3Enabled"`     // 是否启用HTTP/3
	OcspIsOn         uint8    `field:"ocspIsOn"`         // 是否启用OCSP
	State            uint8    `field:"state"`            // 状态
	CreatedAt        uint64   `field:"createdAt"`        // 创建时间
}

type SSLPolicyOperator struct {
	Id               any // ID
	AdminId          any // 管理员ID
	UserId           any // 用户ID
	IsOn             any // 是否启用
	Certs            any // 证书列表
	ClientCACerts    any // 客户端证书
	ClientAuthType   any // 客户端认证类型
	MinVersion       any // 支持的SSL最小版本
	CipherSuitesIsOn any // 是否自定义加密算法套件
	CipherSuites     any // 加密算法套件
	Hsts             any // HSTS设置
	Http2Enabled     any // 是否启用HTTP/2
	Http3Enabled     any // 是否启用HTTP/3
	OcspIsOn         any // 是否启用OCSP
	State            any // 状态
	CreatedAt        any // 创建时间
}

func NewSSLPolicyOperator() *SSLPolicyOperator {
	return &SSLPolicyOperator{}
}
