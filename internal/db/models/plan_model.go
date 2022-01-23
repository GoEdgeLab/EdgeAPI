package models

// Plan 用户套餐
type Plan struct {
	Id              uint32  `field:"id"`              // ID
	IsOn            uint8   `field:"isOn"`            // 是否启用
	Name            string  `field:"name"`            // 套餐名
	ClusterId       uint32  `field:"clusterId"`       // 集群ID
	TrafficLimit    string  `field:"trafficLimit"`    // 流量限制
	Features        string  `field:"features"`        // 允许的功能
	TrafficPrice    string  `field:"trafficPrice"`    // 流量价格设定
	BandwidthPrice  string  `field:"bandwidthPrice"`  // 带宽价格
	MonthlyPrice    float64 `field:"monthlyPrice"`    // 月付
	SeasonallyPrice float64 `field:"seasonallyPrice"` // 季付
	YearlyPrice     float64 `field:"yearlyPrice"`     // 年付
	PriceType       string  `field:"priceType"`       // 价格类型
	Order           uint32  `field:"order"`           // 排序
	State           uint8   `field:"state"`           // 状态
}

type PlanOperator struct {
	Id              interface{} // ID
	IsOn            interface{} // 是否启用
	Name            interface{} // 套餐名
	ClusterId       interface{} // 集群ID
	TrafficLimit    interface{} // 流量限制
	Features        interface{} // 允许的功能
	TrafficPrice    interface{} // 流量价格设定
	BandwidthPrice  interface{} // 带宽价格
	MonthlyPrice    interface{} // 月付
	SeasonallyPrice interface{} // 季付
	YearlyPrice     interface{} // 年付
	PriceType       interface{} // 价格类型
	Order           interface{} // 排序
	State           interface{} // 状态
}

func NewPlanOperator() *PlanOperator {
	return &PlanOperator{}
}
