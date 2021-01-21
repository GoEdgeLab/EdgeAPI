package stats

// 总的流量统计（按天）
type NodeClusterTrafficDailyStat struct {
	Id        uint64 `field:"id"`        // ID
	ClusterId uint32 `field:"clusterId"` // 集群ID
	Day       string `field:"day"`       // YYYYMMDD
	Bytes     uint64 `field:"bytes"`     // 流量字节
}

type NodeClusterTrafficDailyStatOperator struct {
	Id        interface{} // ID
	ClusterId interface{} // 集群ID
	Day       interface{} // YYYYMMDD
	Bytes     interface{} // 流量字节
}

func NewNodeClusterTrafficDailyStatOperator() *NodeClusterTrafficDailyStatOperator {
	return &NodeClusterTrafficDailyStatOperator{}
}
