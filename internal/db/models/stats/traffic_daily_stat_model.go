package stats

// 总的流量统计
type TrafficDailyStat struct {
	Id    uint64 `field:"id"`    // ID
	Day   string `field:"day"`   // YYYYMMDD
	Bytes uint64 `field:"bytes"` // 流量字节
}

type TrafficDailyStatOperator struct {
	Id    interface{} // ID
	Day   interface{} // YYYYMMDD
	Bytes interface{} // 流量字节
}

func NewTrafficDailyStatOperator() *TrafficDailyStatOperator {
	return &TrafficDailyStatOperator{}
}
