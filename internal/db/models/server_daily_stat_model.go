package models

// 计费流量统计
type ServerDailyStat struct {
	Id       uint64 `field:"id"`       // ID
	ServerId uint32 `field:"serverId"` // 服务ID
	RegionId uint32 `field:"regionId"` // 区域ID
	Bytes    uint64 `field:"bytes"`    // 流量
	Day      string `field:"day"`      // 日期YYYYMMDD
	TimeFrom string `field:"timeFrom"` // 开始时间HHMMSS
	TimeTo   string `field:"timeTo"`   // 结束时间
}

type ServerDailyStatOperator struct {
	Id       interface{} // ID
	ServerId interface{} // 服务ID
	RegionId interface{} // 区域ID
	Bytes    interface{} // 流量
	Day      interface{} // 日期YYYYMMDD
	TimeFrom interface{} // 开始时间HHMMSS
	TimeTo   interface{} // 结束时间
}

func NewServerDailyStatOperator() *ServerDailyStatOperator {
	return &ServerDailyStatOperator{}
}
