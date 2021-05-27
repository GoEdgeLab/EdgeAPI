package nameservers

// NSRoute DNS线路
type NSRoute struct {
	Id        uint32 `field:"id"`        // ID
	ClusterId uint32 `field:"clusterId"` // 集群ID
	DomainId  uint32 `field:"domainId"`  // 域名ID
	UserId    uint32 `field:"userId"`    // 用户ID
	Name      string `field:"name"`      // 名称
	Conds     string `field:"conds"`     // 条件定义
	IsOn      uint8  `field:"isOn"`      // 是否启用
	State     uint8  `field:"state"`     // 状态
}

type NSRouteOperator struct {
	Id        interface{} // ID
	ClusterId interface{} // 集群ID
	DomainId  interface{} // 域名ID
	UserId    interface{} // 用户ID
	Name      interface{} // 名称
	Conds     interface{} // 条件定义
	IsOn      interface{} // 是否启用
	State     interface{} // 状态
}

func NewNSRouteOperator() *NSRouteOperator {
	return &NSRouteOperator{}
}
