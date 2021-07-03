package models

// MetricChart 指标图表
type MetricChart struct {
	Id       uint32 `field:"id"`       // ID
	ItemId   uint32 `field:"itemId"`   // 指标ID
	Name     string `field:"name"`     // 名称
	Type     string `field:"type"`     // 图形类型
	WidthDiv int32  `field:"widthDiv"` // 宽度划分
	Params   string `field:"params"`   // 图形参数
	Order    uint32 `field:"order"`    // 排序
	IsOn     uint8  `field:"isOn"`     // 是否启用
	State    uint8  `field:"state"`    // 状态
	MaxItems uint32 `field:"maxItems"` // 最多条目
}

type MetricChartOperator struct {
	Id       interface{} // ID
	ItemId   interface{} // 指标ID
	Name     interface{} // 名称
	Type     interface{} // 图形类型
	WidthDiv interface{} // 宽度划分
	Params   interface{} // 图形参数
	Order    interface{} // 排序
	IsOn     interface{} // 是否启用
	State    interface{} // 状态
	MaxItems interface{} // 最多条目
}

func NewMetricChartOperator() *MetricChartOperator {
	return &MetricChartOperator{}
}
