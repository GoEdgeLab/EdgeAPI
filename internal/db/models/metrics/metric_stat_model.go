package metrics

// MetricStat 指标统计数据
type MetricStat struct {
	Id        uint64  `field:"id"`        // ID
	ClusterId uint32  `field:"clusterId"` // 集群ID
	NodeId    uint32  `field:"nodeId"`    // 节点ID
	ServerId  uint32  `field:"serverId"`  // 服务ID
	ItemId    uint64  `field:"itemId"`    // 指标
	KeyId     uint64  `field:"keyId"`     // 指标键ID
	Value     float64 `field:"value"`     // 数值
	Minute    string  `field:"minute"`    // 分钟值YYYYMMDDHHII
}

type MetricStatOperator struct {
	Id        interface{} // ID
	ClusterId interface{} // 集群ID
	NodeId    interface{} // 节点ID
	ServerId  interface{} // 服务ID
	ItemId    interface{} // 指标
	KeyId     interface{} // 指标键ID
	Value     interface{} // 数值
	Minute    interface{} // 分钟值YYYYMMDDHHII
}

func NewMetricStatOperator() *MetricStatOperator {
	return &MetricStatOperator{}
}
