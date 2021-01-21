package stats

// 总的流量统计（按小时）
type TrafficHourlyStat struct {
	Id    uint64 `field:"id"`    // ID
	Hour  string `field:"hour"`  // YYYYMMDDHH
	Bytes uint64 `field:"bytes"` // 流量字节
}

type TrafficHourlyStatOperator struct {
	Id    interface{} // ID
	Hour  interface{} // YYYYMMDDHH
	Bytes interface{} // 流量字节
}

func NewTrafficHourlyStatOperator() *TrafficHourlyStatOperator {
	return &TrafficHourlyStatOperator{}
}
