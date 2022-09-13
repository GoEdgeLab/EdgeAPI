package nameservers

import "github.com/iwind/TeaGo/dbs"

// NSPlan NS套餐
type NSPlan struct {
	Id           uint32   `field:"id"`           // ID
	Name         string   `field:"name"`         // 套餐名称
	IsOn         bool     `field:"isOn"`         // 是否启用
	MonthlyPrice float64  `field:"monthlyPrice"` // 月价格
	YearlyPrice  float64  `field:"yearlyPrice"`  // 年价格
	Order        uint32   `field:"order"`        // 排序
	Config       dbs.JSON `field:"config"`       // 配置
	State        uint8    `field:"state"`        // 状态
}

type NSPlanOperator struct {
	Id           any // ID
	Name         any // 套餐名称
	IsOn         any // 是否启用
	MonthlyPrice any // 月价格
	YearlyPrice  any // 年价格
	Order        any // 排序
	Config       any // 配置
	State        any // 状态
}

func NewNSPlanOperator() *NSPlanOperator {
	return &NSPlanOperator{}
}
