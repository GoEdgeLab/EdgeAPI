package models

// ADPackagePrice 流量包价格
type ADPackagePrice struct {
	Id            uint32  `field:"id"`            // ID
	PackageId     uint32  `field:"packageId"`     // 高防产品ID
	PeriodId      uint32  `field:"periodId"`      // 有效期ID
	Price         float64 `field:"price"`         // 价格
	DiscountPrice float64 `field:"discountPrice"` // 折后价格
}

type ADPackagePriceOperator struct {
	Id            any // ID
	PackageId     any // 高防产品ID
	PeriodId      any // 有效期ID
	Price         any // 价格
	DiscountPrice any // 折后价格
}

func NewADPackagePriceOperator() *ADPackagePriceOperator {
	return &ADPackagePriceOperator{}
}
