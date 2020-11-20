package models

// 操作日志
type Log struct {
	Id          uint32 `field:"id"`          // ID
	Level       string `field:"level"`       // 级别
	Description string `field:"description"` // 描述
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	Action      string `field:"action"`      // 动作
	UserId      uint32 `field:"userId"`      // 用户ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	ProviderId  uint32 `field:"providerId"`  // 供应商ID
	Ip          string `field:"ip"`          // IP地址
	Type        string `field:"type"`        // 类型：admin, user
	Day         string `field:"day"`         // 日期
}

type LogOperator struct {
	Id          interface{} // ID
	Level       interface{} // 级别
	Description interface{} // 描述
	CreatedAt   interface{} // 创建时间
	Action      interface{} // 动作
	UserId      interface{} // 用户ID
	AdminId     interface{} // 管理员ID
	ProviderId  interface{} // 供应商ID
	Ip          interface{} // IP地址
	Type        interface{} // 类型：admin, user
	Day         interface{} // 日期
}

func NewLogOperator() *LogOperator {
	return &LogOperator{}
}
