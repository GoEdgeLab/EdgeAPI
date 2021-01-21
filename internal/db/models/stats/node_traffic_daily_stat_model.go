package stats

// 总的流量统计（按天）
type NodeTrafficDailyStat struct {
	Id     uint64 `field:"id"`     // ID
	NodeId uint32 `field:"nodeId"` // 集群ID
	Day    string `field:"day"`    // YYYYMMDD
	Bytes  uint64 `field:"bytes"`  // 流量字节
}

type NodeTrafficDailyStatOperator struct {
	Id     interface{} // ID
	NodeId interface{} // 集群ID
	Day    interface{} // YYYYMMDD
	Bytes  interface{} // 流量字节
}

func NewNodeTrafficDailyStatOperator() *NodeTrafficDailyStatOperator {
	return &NodeTrafficDailyStatOperator{}
}
