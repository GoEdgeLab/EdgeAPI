package nameservers

import "github.com/iwind/TeaGo/dbs"

// NSDomain DNS域名
type NSDomain struct {
	Id                 uint64   `field:"id"`                 // ID
	ClusterId          uint32   `field:"clusterId"`          // 集群ID
	UserId             uint32   `field:"userId"`             // 用户ID
	IsOn               bool     `field:"isOn"`               // 是否启用
	Name               string   `field:"name"`               // 域名
	GroupIds           dbs.JSON `field:"groupIds"`           // 分组ID
	Tsig               dbs.JSON `field:"tsig"`               // TSIG配置
	VerifyTXT          string   `field:"verifyTXT"`          // 验证用的TXT
	VerifyExpiresAt    uint64   `field:"verifyExpiresAt"`    // 验证TXT过期时间
	RecordsHealthCheck dbs.JSON `field:"recordsHealthCheck"` // 记录健康检查设置
	CreatedAt          uint64   `field:"createdAt"`          // 创建时间
	Version            uint64   `field:"version"`            // 版本号
	Status             string   `field:"status"`             // 状态：none|verified
	State              uint8    `field:"state"`              // 状态
}

type NSDomainOperator struct {
	Id                 any // ID
	ClusterId          any // 集群ID
	UserId             any // 用户ID
	IsOn               any // 是否启用
	Name               any // 域名
	GroupIds           any // 分组ID
	Tsig               any // TSIG配置
	VerifyTXT          any // 验证用的TXT
	VerifyExpiresAt    any // 验证TXT过期时间
	RecordsHealthCheck any // 记录健康检查设置
	CreatedAt          any // 创建时间
	Version            any // 版本号
	Status             any // 状态：none|verified
	State              any // 状态
}

func NewNSDomainOperator() *NSDomainOperator {
	return &NSDomainOperator{}
}
