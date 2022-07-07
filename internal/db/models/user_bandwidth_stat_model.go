package models

// UserBandwidthStat 用户月带宽峰值
type UserBandwidthStat struct {
	Id     uint64 `field:"id"`     // ID
	UserId uint64 `field:"userId"` // 用户ID
	Day    string `field:"day"`    // 日期YYYYMMDD
	TimeAt string `field:"timeAt"` // 时间点HHII
	Bytes  uint64 `field:"bytes"`  // 带宽
}

type UserBandwidthStatOperator struct {
	Id     interface{} // ID
	UserId interface{} // 用户ID
	Day    interface{} // 日期YYYYMMDD
	TimeAt interface{} // 时间点HHII
	Bytes  interface{} // 带宽
}

func NewUserBandwidthStatOperator() *UserBandwidthStatOperator {
	return &UserBandwidthStatOperator{}
}
