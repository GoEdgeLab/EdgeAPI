package nameservers

import "github.com/iwind/TeaGo/dbs"

// NSDomain DNS域名
type NSDomain struct {
	Id        uint32   `field:"id"`        // ID
	ClusterId uint32   `field:"clusterId"` // 集群ID
	UserId    uint32   `field:"userId"`    // 用户ID
	IsOn      uint8    `field:"isOn"`      // 是否启用
	Name      string   `field:"name"`      // 域名
	CreatedAt uint64   `field:"createdAt"` // 创建时间
	Version   uint64   `field:"version"`   // 版本
	State     uint8    `field:"state"`     // 状态
	Tsig      dbs.JSON `field:"tsig"`      // TSIG配置
}

type NSDomainOperator struct {
	Id        interface{} // ID
	ClusterId interface{} // 集群ID
	UserId    interface{} // 用户ID
	IsOn      interface{} // 是否启用
	Name      interface{} // 域名
	CreatedAt interface{} // 创建时间
	Version   interface{} // 版本
	State     interface{} // 状态
	Tsig      interface{} // TSIG配置
}

func NewNSDomainOperator() *NSDomainOperator {
	return &NSDomainOperator{}
}
