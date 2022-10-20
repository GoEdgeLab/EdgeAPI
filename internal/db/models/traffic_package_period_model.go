package models

// TrafficPackagePeriod 流量包有效期
type TrafficPackagePeriod struct {
	Id     uint32 `field:"id"`     // ID
	IsOn   bool   `field:"isOn"`   // 是否启用
	Count  uint32 `field:"count"`  // 数量
	Unit   string `field:"unit"`   // 单位：month, year
	Months uint32 `field:"months"` // 月数
	State  uint8  `field:"state"`  // 状态
}

type TrafficPackagePeriodOperator struct {
	Id     any // ID
	IsOn   any // 是否启用
	Count  any // 数量
	Unit   any // 单位：month, year
	Months any // 月数
	State  any // 状态
}

func NewTrafficPackagePeriodOperator() *TrafficPackagePeriodOperator {
	return &TrafficPackagePeriodOperator{}
}
