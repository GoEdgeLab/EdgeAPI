package models

// ServerDailyStat 计费流量统计
type ServerDailyStat struct {
	Id                  uint64  `field:"id"`                  // ID
	UserId              uint32  `field:"userId"`              // 用户ID
	ServerId            uint32  `field:"serverId"`            // 服务ID
	RegionId            uint32  `field:"regionId"`            // 区域ID
	Bytes               uint64  `field:"bytes"`               // 流量
	CachedBytes         uint64  `field:"cachedBytes"`         // 缓存的流量
	CountRequests       uint64  `field:"countRequests"`       // 请求数
	CountCachedRequests uint64  `field:"countCachedRequests"` // 缓存的请求数
	CountAttackRequests uint64  `field:"countAttackRequests"` // 攻击请求数
	AttackBytes         uint64  `field:"attackBytes"`         // 攻击流量
	Day                 string  `field:"day"`                 // 日期YYYYMMDD
	Hour                string  `field:"hour"`                // YYYYMMDDHH
	TimeFrom            string  `field:"timeFrom"`            // 开始时间HHMMSS
	TimeTo              string  `field:"timeTo"`              // 结束时间
	IsCharged           bool    `field:"isCharged"`           // 是否已计算费用
	PlanId              uint64  `field:"planId"`              // 套餐ID
	Fee                 float64 `field:"fee"`                 // 费用
}

type ServerDailyStatOperator struct {
	Id                  interface{} // ID
	UserId              interface{} // 用户ID
	ServerId            interface{} // 服务ID
	RegionId            interface{} // 区域ID
	Bytes               interface{} // 流量
	CachedBytes         interface{} // 缓存的流量
	CountRequests       interface{} // 请求数
	CountCachedRequests interface{} // 缓存的请求数
	CountAttackRequests interface{} // 攻击请求数
	AttackBytes         interface{} // 攻击流量
	Day                 interface{} // 日期YYYYMMDD
	Hour                interface{} // YYYYMMDDHH
	TimeFrom            interface{} // 开始时间HHMMSS
	TimeTo              interface{} // 结束时间
	IsCharged           interface{} // 是否已计算费用
	PlanId              interface{} // 套餐ID
	Fee                 interface{} // 费用
}

func NewServerDailyStatOperator() *ServerDailyStatOperator {
	return &ServerDailyStatOperator{}
}
