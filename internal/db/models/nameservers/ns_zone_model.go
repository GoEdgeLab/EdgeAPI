package nameservers

// NSZone 域名子域
type NSZone struct {
	Id       uint64 `field:"id"`       // ID
	DomainId uint64 `field:"domainId"` // 域名ID
	IsOn     uint8  `field:"isOn"`     // 是否启用
	Order    uint32 `field:"order"`    // 排序
	Version  uint64 `field:"version"`  // 版本
	Tsig     string `field:"tsig"`     // TSIG配置
	State    uint8  `field:"state"`    // 状态
}

type NSZoneOperator struct {
	Id       interface{} // ID
	DomainId interface{} // 域名ID
	IsOn     interface{} // 是否启用
	Order    interface{} // 排序
	Version  interface{} // 版本
	Tsig     interface{} // TSIG配置
	State    interface{} // 状态
}

func NewNSZoneOperator() *NSZoneOperator {
	return &NSZoneOperator{}
}
