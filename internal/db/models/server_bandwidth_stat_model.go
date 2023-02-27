package models

// ServerBandwidthStat 服务峰值带宽统计
type ServerBandwidthStat struct {
	Id         uint64 `field:"id"`         // ID
	UserId     uint64 `field:"userId"`     // 用户ID
	ServerId   uint64 `field:"serverId"`   // 服务ID
	RegionId   uint32 `field:"regionId"`   // 区域ID
	Day        string `field:"day"`        // 日期YYYYMMDD
	TimeAt     string `field:"timeAt"`     // 时间点HHMM
	Bytes      uint64 `field:"bytes"`      // 带宽字节
	AvgBytes   uint64 `field:"avgBytes"`   // 平均流量
	TotalBytes uint64 `field:"totalBytes"` // 总流量
}

type ServerBandwidthStatOperator struct {
	Id         any // ID
	UserId     any // 用户ID
	ServerId   any // 服务ID
	RegionId   any // 区域ID
	Day        any // 日期YYYYMMDD
	TimeAt     any // 时间点HHMM
	Bytes      any // 带宽字节
	AvgBytes   any // 平均流量
	TotalBytes any // 总流量
}

func NewServerBandwidthStatOperator() *ServerBandwidthStatOperator {
	return &ServerBandwidthStatOperator{}
}
