package models

// UserTrafficBill 用户流量/带宽账单
type UserTrafficBill struct {
	Id           uint64  `field:"id"`           // ID
	BillId       uint64  `field:"billId"`       // 主账单ID
	RegionId     uint32  `field:"regionId"`     // 区域ID
	Amount       float64 `field:"amount"`       // 金额
	BandwidthMB  float64 `field:"bandwidthMB"`  // 带宽MB
	TrafficGB    float64 `field:"trafficGB"`    // 流量GB
	PricePerUnit float64 `field:"pricePerUnit"` // 单位价格
	PriceType    string  `field:"priceType"`    // 计费方式：traffic|bandwidth
	State        uint8   `field:"state"`        // 状态
}

type UserTrafficBillOperator struct {
	Id           any // ID
	BillId       any // 主账单ID
	RegionId     any // 区域ID
	Amount       any // 金额
	BandwidthMB  any // 带宽MB
	TrafficGB    any // 流量GB
	PricePerUnit any // 单位价格
	PriceType    any // 计费方式：traffic|bandwidth
	State        any // 状态
}

func NewUserTrafficBillOperator() *UserTrafficBillOperator {
	return &UserTrafficBillOperator{}
}
