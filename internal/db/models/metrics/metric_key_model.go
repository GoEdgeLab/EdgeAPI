package metrics

// MetricKey 指标键值
type MetricKey struct {
	Id     uint64 `field:"id"`     // ID
	ItemId uint64 `field:"itemId"` // 指标ID
	Value  string `field:"value"`  // 值
	Hash   string `field:"hash"`   // 对值进行Hash
}

type MetricKeyOperator struct {
	Id     interface{} // ID
	ItemId interface{} // 指标ID
	Value  interface{} // 值
	Hash   interface{} // 对值进行Hash
}

func NewMetricKeyOperator() *MetricKeyOperator {
	return &MetricKeyOperator{}
}
