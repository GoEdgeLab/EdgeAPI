package models

import "github.com/iwind/TeaGo/dbs"

// UserADInstance 高防实例
type UserADInstance struct {
	Id          uint64   `field:"id"`          // ID
	AdminId     uint32   `field:"adminId"`     // 管理员ID
	UserId      uint64   `field:"userId"`      // 用户ID
	InstanceId  uint32   `field:"instanceId"`  // 高防实例ID
	PeriodId    uint32   `field:"periodId"`    // 有效期
	PeriodCount uint32   `field:"periodCount"` // 有效期数量
	PeriodUnit  string   `field:"periodUnit"`  // 有效期单位
	DayFrom     string   `field:"dayFrom"`     // 开始日期
	DayTo       string   `field:"dayTo"`       // 结束日期
	MaxObjects  uint32   `field:"maxObjects"`  // 最多防护对象数
	ObjectCodes dbs.JSON `field:"objectCodes"` // 防护对象
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
	State       uint8    `field:"state"`       // 状态
}

type UserADInstanceOperator struct {
	Id          any // ID
	AdminId     any // 管理员ID
	UserId      any // 用户ID
	InstanceId  any // 高防实例ID
	PeriodId    any // 有效期
	PeriodCount any // 有效期数量
	PeriodUnit  any // 有效期单位
	DayFrom     any // 开始日期
	DayTo       any // 结束日期
	MaxObjects  any // 最多防护对象数
	ObjectCodes any // 防护对象
	CreatedAt   any // 创建时间
	State       any // 状态
}

func NewUserADInstanceOperator() *UserADInstanceOperator {
	return &UserADInstanceOperator{}
}
