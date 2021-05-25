package nameservers

// NSDomain DNS域名
type NSDomain struct {
	Id        uint32 `field:"id"`        // ID
	ClusterId uint32 `field:"clusterId"` // 集群ID
	UserId    uint32 `field:"userId"`    // 用户ID
	IsOn      uint8  `field:"isOn"`      // 是否启用
	Name      string `field:"name"`      // 域名
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	State     uint8  `field:"state"`     // 状态
}

type NSDomainOperator struct {
	Id        interface{} // ID
	ClusterId interface{} // 集群ID
	UserId    interface{} // 用户ID
	IsOn      interface{} // 是否启用
	Name      interface{} // 域名
	CreatedAt interface{} // 创建时间
	State     interface{} // 状态
}

func NewNSDomainOperator() *NSDomainOperator {
	return &NSDomainOperator{}
}
