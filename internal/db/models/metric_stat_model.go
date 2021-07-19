package models

// MetricStat 指标统计数据
type MetricStat struct {
	Id         uint64  `field:"id"`         // ID
	Hash       string  `field:"hash"`       // Hash值
	ClusterId  uint32  `field:"clusterId"`  // 集群ID
	NodeId     uint32  `field:"nodeId"`     // 节点ID
	ServerId   uint32  `field:"serverId"`   // 服务ID
	ItemId     uint64  `field:"itemId"`     // 指标
	Keys       string  `field:"keys"`       // 键值
	Value      float64 `field:"value"`      // 数值
	Time       string  `field:"time"`       // 分钟值YYYYMMDDHHII
	Version    uint32  `field:"version"`    // 版本号
	CreatedDay string  `field:"createdDay"` // YYYYMMDD
}

type MetricStatOperator struct {
	Id         interface{} // ID
	Hash       interface{} // Hash值
	ClusterId  interface{} // 集群ID
	NodeId     interface{} // 节点ID
	ServerId   interface{} // 服务ID
	ItemId     interface{} // 指标
	Keys       interface{} // 键值
	Value      interface{} // 数值
	Time       interface{} // 分钟值YYYYMMDDHHII
	Version    interface{} // 版本号
	CreatedDay interface{} // YYYYMMDD
}

func NewMetricStatOperator() *MetricStatOperator {
	return &MetricStatOperator{}
}
