package nameservers

import "github.com/iwind/TeaGo/dbs"

// NSDomain DNS域名
type NSDomain struct {
	Id        uint64   `field:"id"`        // ID
	ClusterId uint32   `field:"clusterId"` // 集群ID
	UserId    uint32   `field:"userId"`    // 用户ID
	IsOn      bool     `field:"isOn"`      // 是否启用
	Name      string   `field:"name"`      // 域名
	GroupIds  dbs.JSON `field:"groupIds"`  // 分组ID
	Tsig      dbs.JSON `field:"tsig"`      // TSIG配置
	CreatedAt uint64   `field:"createdAt"` // 创建时间
	Version   uint64   `field:"version"`   // 版本号
	Status    string   `field:"status"`    // 状态：none|verified
	State     uint8    `field:"state"`     // 状态
}

type NSDomainOperator struct {
	Id        interface{} // ID
	ClusterId interface{} // 集群ID
	UserId    interface{} // 用户ID
	IsOn      interface{} // 是否启用
	Name      interface{} // 域名
	GroupIds  interface{} // 分组ID
	Tsig      interface{} // TSIG配置
	CreatedAt interface{} // 创建时间
	Version   interface{} // 版本号
	Status    interface{} // 状态：none|verified
	State     interface{} // 状态
}

func NewNSDomainOperator() *NSDomainOperator {
	return &NSDomainOperator{}
}
