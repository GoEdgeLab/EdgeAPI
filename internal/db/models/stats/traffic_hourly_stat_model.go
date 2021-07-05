package stats

// TrafficHourlyStat 总的流量统计（按小时）
type TrafficHourlyStat struct {
	Id                  uint64 `field:"id"`                  // ID
	Hour                string `field:"hour"`                // YYYYMMDDHH
	Bytes               uint64 `field:"bytes"`               // 流量字节
	CachedBytes         uint64 `field:"cachedBytes"`         // 缓存流量
	CountRequests       uint64 `field:"countRequests"`       // 请求数
	CountCachedRequests uint64 `field:"countCachedRequests"` // 缓存请求数
}

type TrafficHourlyStatOperator struct {
	Id                  interface{} // ID
	Hour                interface{} // YYYYMMDDHH
	Bytes               interface{} // 流量字节
	CachedBytes         interface{} // 缓存流量
	CountRequests       interface{} // 请求数
	CountCachedRequests interface{} // 缓存请求数
}

func NewTrafficHourlyStatOperator() *TrafficHourlyStatOperator {
	return &TrafficHourlyStatOperator{}
}
