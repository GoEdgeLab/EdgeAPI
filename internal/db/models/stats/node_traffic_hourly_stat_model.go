package stats

// NodeTrafficHourlyStat 总的流量统计（按天）
type NodeTrafficHourlyStat struct {
	Id                  uint64 `field:"id"`                  // ID
	Role                string `field:"role"`                // 节点角色
	ClusterId           uint32 `field:"clusterId"`           // 集群ID
	NodeId              uint32 `field:"nodeId"`              // 集群ID
	Hour                string `field:"hour"`                // YYYYMMDDHH
	Bytes               uint64 `field:"bytes"`               // 流量字节
	CachedBytes         uint64 `field:"cachedBytes"`         // 缓存流量
	CountRequests       uint64 `field:"countRequests"`       // 请求数
	CountCachedRequests uint64 `field:"countCachedRequests"` // 缓存的请求数
	CountAttackRequests uint64 `field:"countAttackRequests"` // 攻击请求数
	AttackBytes         uint64 `field:"attackBytes"`         // 攻击流量
}

type NodeTrafficHourlyStatOperator struct {
	Id                  interface{} // ID
	Role                interface{} // 节点角色
	ClusterId           interface{} // 集群ID
	NodeId              interface{} // 集群ID
	Hour                interface{} // YYYYMMDDHH
	Bytes               interface{} // 流量字节
	CachedBytes         interface{} // 缓存流量
	CountRequests       interface{} // 请求数
	CountCachedRequests interface{} // 缓存的请求数
	CountAttackRequests interface{} // 攻击请求数
	AttackBytes         interface{} // 攻击流量
}

func NewNodeTrafficHourlyStatOperator() *NodeTrafficHourlyStatOperator {
	return &NodeTrafficHourlyStatOperator{}
}
