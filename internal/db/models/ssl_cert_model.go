package models

import "github.com/iwind/TeaGo/dbs"

// SSLCert SSL证书
type SSLCert struct {
	Id                 uint32   `field:"id"`                 // ID
	AdminId            uint32   `field:"adminId"`            // 管理员ID
	UserId             uint32   `field:"userId"`             // 用户ID
	State              uint8    `field:"state"`              // 状态
	CreatedAt          uint64   `field:"createdAt"`          // 创建时间
	UpdatedAt          uint64   `field:"updatedAt"`          // 修改时间
	IsOn               uint8    `field:"isOn"`               // 是否启用
	Name               string   `field:"name"`               // 证书名
	Description        string   `field:"description"`        // 描述
	CertData           []byte   `field:"certData"`           // 证书内容
	KeyData            []byte   `field:"keyData"`            // 密钥内容
	ServerName         string   `field:"serverName"`         // 证书使用的主机名
	IsCA               uint8    `field:"isCA"`               // 是否为CA证书
	GroupIds           dbs.JSON `field:"groupIds"`           // 证书分组
	TimeBeginAt        uint64   `field:"timeBeginAt"`        // 开始时间
	TimeEndAt          uint64   `field:"timeEndAt"`          // 结束时间
	DnsNames           dbs.JSON `field:"dnsNames"`           // DNS名称列表
	CommonNames        dbs.JSON `field:"commonNames"`        // 发行单位列表
	IsACME             uint8    `field:"isACME"`             // 是否为ACME自动生成的
	AcmeTaskId         uint64   `field:"acmeTaskId"`         // ACME任务ID
	NotifiedAt         uint64   `field:"notifiedAt"`         // 最后通知时间
	Ocsp               []byte   `field:"ocsp"`               // OCSP缓存
	OcspIsUpdated      uint8    `field:"ocspIsUpdated"`      // OCSP是否已更新
	OcspUpdatedAt      uint64   `field:"ocspUpdatedAt"`      // OCSP更新时间
	OcspError          string   `field:"ocspError"`          // OCSP更新错误
	OcspUpdatedVersion uint64   `field:"ocspUpdatedVersion"` // OCSP更新版本
	OcspExpiresAt      uint64   `field:"ocspExpiresAt"`      // OCSP过期时间(UTC)
	OcspTries          uint32   `field:"ocspTries"`          // OCSP尝试次数
}

type SSLCertOperator struct {
	Id                 interface{} // ID
	AdminId            interface{} // 管理员ID
	UserId             interface{} // 用户ID
	State              interface{} // 状态
	CreatedAt          interface{} // 创建时间
	UpdatedAt          interface{} // 修改时间
	IsOn               interface{} // 是否启用
	Name               interface{} // 证书名
	Description        interface{} // 描述
	CertData           interface{} // 证书内容
	KeyData            interface{} // 密钥内容
	ServerName         interface{} // 证书使用的主机名
	IsCA               interface{} // 是否为CA证书
	GroupIds           interface{} // 证书分组
	TimeBeginAt        interface{} // 开始时间
	TimeEndAt          interface{} // 结束时间
	DnsNames           interface{} // DNS名称列表
	CommonNames        interface{} // 发行单位列表
	IsACME             interface{} // 是否为ACME自动生成的
	AcmeTaskId         interface{} // ACME任务ID
	NotifiedAt         interface{} // 最后通知时间
	Ocsp               interface{} // OCSP缓存
	OcspIsUpdated      interface{} // OCSP是否已更新
	OcspUpdatedAt      interface{} // OCSP更新时间
	OcspError          interface{} // OCSP更新错误
	OcspUpdatedVersion interface{} // OCSP更新版本
	OcspExpiresAt      interface{} // OCSP过期时间(UTC)
	OcspTries          interface{} // OCSP尝试次数
}

func NewSSLCertOperator() *SSLCertOperator {
	return &SSLCertOperator{}
}
