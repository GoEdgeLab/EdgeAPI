package models

// ServerBandwidthStat 服务峰值带宽统计
type ServerBandwidthStat struct {
	Id       uint64 `field:"id"`       // ID
	UserId   uint64 `field:"userId"`   // 用户ID
	ServerId uint64 `field:"serverId"` // 服务ID
	Day      string `field:"day"`      // 日期YYYYMMDD
	TimeAt   string `field:"timeAt"`   // 时间点HHMM
	Bytes    uint64 `field:"bytes"`    // 带宽字节
}

type ServerBandwidthStatOperator struct {
	Id       interface{} // ID
	UserId   interface{} // 用户ID
	ServerId interface{} // 服务ID
	Day      interface{} // 日期YYYYMMDD
	TimeAt   interface{} // 时间点HHMM
	Bytes    interface{} // 带宽字节
}

func NewServerBandwidthStatOperator() *ServerBandwidthStatOperator {
	return &ServerBandwidthStatOperator{}
}
