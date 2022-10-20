package models

// TrafficPackagePrice 流量包价格
type TrafficPackagePrice struct {
	Id            uint32  `field:"id"`            // ID
	PackageId     uint32  `field:"packageId"`     // 套餐ID
	RegionId      uint32  `field:"regionId"`      // 区域ID
	PeriodId      uint32  `field:"periodId"`      // 有效期ID
	Price         float64 `field:"price"`         // 价格
	DiscountPrice float64 `field:"discountPrice"` // 折后价格
}

type TrafficPackagePriceOperator struct {
	Id            any // ID
	PackageId     any // 套餐ID
	RegionId      any // 区域ID
	PeriodId      any // 有效期ID
	Price         any // 价格
	DiscountPrice any // 折后价格
}

func NewTrafficPackagePriceOperator() *TrafficPackagePriceOperator {
	return &TrafficPackagePriceOperator{}
}
