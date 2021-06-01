package nameservers

// NSRoute DNS线路
type NSRoute struct {
	Id        uint32 `field:"id"`        // ID
	IsOn      uint8  `field:"isOn"`      // 是否启用
	ClusterId uint32 `field:"clusterId"` // 集群ID
	DomainId  uint32 `field:"domainId"`  // 域名ID
	UserId    uint32 `field:"userId"`    // 用户ID
	Name      string `field:"name"`      // 名称
	Ranges    string `field:"ranges"`    // 范围
	Order     uint32 `field:"order"`     // 排序
	Version   uint64 `field:"version"`   // 版本号
	State     uint8  `field:"state"`     // 状态
}

type NSRouteOperator struct {
	Id        interface{} // ID
	IsOn      interface{} // 是否启用
	ClusterId interface{} // 集群ID
	DomainId  interface{} // 域名ID
	UserId    interface{} // 用户ID
	Name      interface{} // 名称
	Ranges    interface{} // 范围
	Order     interface{} // 排序
	Version   interface{} // 版本号
	State     interface{} // 状态
}

func NewNSRouteOperator() *NSRouteOperator {
	return &NSRouteOperator{}
}
