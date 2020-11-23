package models

// DNS服务商
type DNSProvider struct {
	Id            uint32 `field:"id"`            // ID
	Name          string `field:"name"`          // 名称
	AdminId       uint32 `field:"adminId"`       // 管理员ID
	UserId        uint32 `field:"userId"`        // 用户ID
	Type          string `field:"type"`          // 供应商类型
	ApiParams     string `field:"apiParams"`     // API参数
	CreatedAt     uint64 `field:"createdAt"`     // 创建时间
	State         uint8  `field:"state"`         // 状态
	DataUpdatedAt uint64 `field:"dataUpdatedAt"` // 数据同步时间
}

type DNSProviderOperator struct {
	Id            interface{} // ID
	Name          interface{} // 名称
	AdminId       interface{} // 管理员ID
	UserId        interface{} // 用户ID
	Type          interface{} // 供应商类型
	ApiParams     interface{} // API参数
	CreatedAt     interface{} // 创建时间
	State         interface{} // 状态
	DataUpdatedAt interface{} // 数据同步时间
}

func NewDNSProviderOperator() *DNSProviderOperator {
	return &DNSProviderOperator{}
}
