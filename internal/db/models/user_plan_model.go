package models

// UserPlan 用户的套餐
type UserPlan struct {
	Id     uint64 `field:"id"`     // ID
	UserId uint32 `field:"userId"` // 用户ID
	PlanId uint32 `field:"planId"` // 套餐ID
	IsOn   uint8  `field:"isOn"`   // 是否启用
	Name   string `field:"name"`   // 名称
	DayTo  string `field:"dayTo"`  // 结束日期
	State  uint8  `field:"state"`  // 状态
}

type UserPlanOperator struct {
	Id     interface{} // ID
	UserId interface{} // 用户ID
	PlanId interface{} // 套餐ID
	IsOn   interface{} // 是否启用
	Name   interface{} // 名称
	DayTo  interface{} // 结束日期
	State  interface{} // 状态
}

func NewUserPlanOperator() *UserPlanOperator {
	return &UserPlanOperator{}
}
