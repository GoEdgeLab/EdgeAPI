package models

import "github.com/iwind/TeaGo/dbs"

// UserTrafficBill 用户流量/带宽账单
type UserTrafficBill struct {
	Id                    uint64   `field:"id"`                    // ID
	BillId                uint64   `field:"billId"`                // 主账单ID
	RegionId              uint32   `field:"regionId"`              // 区域ID
	Amount                float64  `field:"amount"`                // 金额
	BandwidthMB           float64  `field:"bandwidthMB"`           // 带宽MB
	BandwidthPercentile   uint8    `field:"bandwidthPercentile"`   // 带宽百分位
	TrafficGB             float64  `field:"trafficGB"`             // 流量GB
	TrafficPackageGB      float64  `field:"trafficPackageGB"`      // 使用的流量包GB
	UserTrafficPackageIds dbs.JSON `field:"userTrafficPackageIds"` // 使用的流量包ID
	PricePerUnit          float64  `field:"pricePerUnit"`          // 单位价格
	PriceType             string   `field:"priceType"`             // 计费方式：traffic|bandwidth
	State                 uint8    `field:"state"`                 // 状态
}

type UserTrafficBillOperator struct {
	Id                    any // ID
	BillId                any // 主账单ID
	RegionId              any // 区域ID
	Amount                any // 金额
	BandwidthMB           any // 带宽MB
	BandwidthPercentile   any // 带宽百分位
	TrafficGB             any // 流量GB
	TrafficPackageGB      any // 使用的流量包GB
	UserTrafficPackageIds any // 使用的流量包ID
	PricePerUnit          any // 单位价格
	PriceType             any // 计费方式：traffic|bandwidth
	State                 any // 状态
}

func NewUserTrafficBillOperator() *UserTrafficBillOperator {
	return &UserTrafficBillOperator{}
}
