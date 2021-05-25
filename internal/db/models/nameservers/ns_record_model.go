package nameservers

// NSRecord DNS记录
type NSRecord struct {
	Id          uint64 `field:"id"`          // ID
	DomainId    uint32 `field:"domainId"`    // 域名ID
	Description string `field:"description"` // 备注
	Name        string `field:"name"`        // 记录名
	Type        string `field:"type"`        // 类型
	Value       string `field:"value"`       // 值
	Ttl         uint32 `field:"ttl"`         // TTL（秒）
	Weight      uint32 `field:"weight"`      // 权重
	Routes      string `field:"routes"`      // 线路
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	State       uint8  `field:"state"`       // 状态
}

type NSRecordOperator struct {
	Id          interface{} // ID
	DomainId    interface{} // 域名ID
	Description interface{} // 备注
	Name        interface{} // 记录名
	Type        interface{} // 类型
	Value       interface{} // 值
	Ttl         interface{} // TTL（秒）
	Weight      interface{} // 权重
	Routes      interface{} // 线路
	CreatedAt   interface{} // 创建时间
	State       interface{} // 状态
}

func NewNSRecordOperator() *NSRecordOperator {
	return &NSRecordOperator{}
}
