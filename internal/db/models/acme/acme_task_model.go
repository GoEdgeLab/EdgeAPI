package acme

import "github.com/iwind/TeaGo/dbs"

// ACMETask ACME任务
type ACMETask struct {
	Id            uint64   `field:"id"`            // ID
	AdminId       uint32   `field:"adminId"`       // 管理员ID
	UserId        uint32   `field:"userId"`        // 用户ID
	IsOn          bool     `field:"isOn"`          // 是否启用
	AcmeUserId    uint32   `field:"acmeUserId"`    // ACME用户ID
	DnsDomain     string   `field:"dnsDomain"`     // DNS主域名
	DnsProviderId uint64   `field:"dnsProviderId"` // DNS服务商
	Domains       dbs.JSON `field:"domains"`       // 证书域名
	CreatedAt     uint64   `field:"createdAt"`     // 创建时间
	State         uint8    `field:"state"`         // 状态
	CertId        uint64   `field:"certId"`        // 生成的证书ID
	AutoRenew     uint8    `field:"autoRenew"`     // 是否自动更新
	AuthType      string   `field:"authType"`      // 认证类型
	AuthURL       string   `field:"authURL"`       // 认证URL
}

type ACMETaskOperator struct {
	Id            interface{} // ID
	AdminId       interface{} // 管理员ID
	UserId        interface{} // 用户ID
	IsOn          interface{} // 是否启用
	AcmeUserId    interface{} // ACME用户ID
	DnsDomain     interface{} // DNS主域名
	DnsProviderId interface{} // DNS服务商
	Domains       interface{} // 证书域名
	CreatedAt     interface{} // 创建时间
	State         interface{} // 状态
	CertId        interface{} // 生成的证书ID
	AutoRenew     interface{} // 是否自动更新
	AuthType      interface{} // 认证类型
	AuthURL       interface{} // 认证URL
}

func NewACMETaskOperator() *ACMETaskOperator {
	return &ACMETaskOperator{}
}
