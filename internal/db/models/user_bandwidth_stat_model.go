package models

// UserBandwidthStat 用户月带宽峰值
type UserBandwidthStat struct {
	Id       uint64 `field:"id"`       // ID
	UserId   uint64 `field:"userId"`   // 用户ID
	Day      string `field:"day"`      // 日期YYYYMMDD
	TimeAt   string `field:"timeAt"`   // 时间点HHII
	Bytes    uint64 `field:"bytes"`    // 带宽
	RegionId uint32 `field:"regionId"` // 区域ID
}

type UserBandwidthStatOperator struct {
	Id       any // ID
	UserId   any // 用户ID
	Day      any // 日期YYYYMMDD
	TimeAt   any // 时间点HHII
	Bytes    any // 带宽
	RegionId any // 区域ID
}

func NewUserBandwidthStatOperator() *UserBandwidthStatOperator {
	return &UserBandwidthStatOperator{}
}
