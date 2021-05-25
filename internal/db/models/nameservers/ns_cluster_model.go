package nameservers

// NSCluster 域名服务器集群
type NSCluster struct {
	Id    uint32 `field:"id"`    // ID
	IsOn  uint8  `field:"isOn"`  // 是否启用
	Name  string `field:"name"`  // 集群名
	State uint8  `field:"state"` // 状态
}

type NSClusterOperator struct {
	Id    interface{} // ID
	IsOn  interface{} // 是否启用
	Name  interface{} // 集群名
	State interface{} // 状态
}

func NewNSClusterOperator() *NSClusterOperator {
	return &NSClusterOperator{}
}
