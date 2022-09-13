package nameservers

// NSUserPlan 用户套餐
type NSUserPlan struct {
	Id         uint64 `field:"id"`         // ID
	UserId     uint64 `field:"userId"`     // 用户ID
	PlanId     uint32 `field:"planId"`     // 套餐ID
	DayFrom    string `field:"dayFrom"`    // YYYYMMDD
	DayTo      string `field:"dayTo"`      // YYYYMMDD
	PeriodUnit string `field:"periodUnit"` // monthly|yearly
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	State      uint8  `field:"state"`      // 状态
}

type NSUserPlanOperator struct {
	Id         any // ID
	UserId     any // 用户ID
	PlanId     any // 套餐ID
	DayFrom    any // YYYYMMDD
	DayTo      any // YYYYMMDD
	PeriodUnit any // monthly|yearly
	CreatedAt  any // 创建时间
	State      any // 状态
}

func NewNSUserPlanOperator() *NSUserPlanOperator {
	return &NSUserPlanOperator{}
}
