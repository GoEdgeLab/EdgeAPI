package nameservers

// NSNode 域名服务器节点
type NSNode struct {
	Id        uint32 `field:"id"`        // ID
	ClusterId uint32 `field:"clusterId"` // 集群ID
	Name      string `field:"name"`      // 节点名称
	IsOn      uint8  `field:"isOn"`      // 是否启用
	Status    string `field:"status"`    // 运行状态
	UniqueId  string `field:"uniqueId"`  // 节点ID
	Secret    string `field:"secret"`    // 密钥
	State     uint8  `field:"state"`     // 状态
}

type NSNodeOperator struct {
	Id        interface{} // ID
	ClusterId interface{} // 集群ID
	Name      interface{} // 节点名称
	IsOn      interface{} // 是否启用
	Status    interface{} // 运行状态
	UniqueId  interface{} // 节点ID
	Secret    interface{} // 密钥
	State     interface{} // 状态
}

func NewNSNodeOperator() *NSNodeOperator {
	return &NSNodeOperator{}
}
