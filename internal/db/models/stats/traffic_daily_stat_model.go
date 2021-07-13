package stats

// TrafficDailyStat 总的流量统计（按天）
type TrafficDailyStat struct {
	Id                  uint64 `field:"id"`                  // ID
	Day                 string `field:"day"`                 // YYYYMMDD
	CachedBytes         uint64 `field:"cachedBytes"`         // 缓存流量
	Bytes               uint64 `field:"bytes"`               // 流量字节
	CountRequests       uint64 `field:"countRequests"`       // 请求数
	CountCachedRequests uint64 `field:"countCachedRequests"` // 缓存请求数
	CountAttackRequests uint64 `field:"countAttackRequests"` // 攻击量
	AttackBytes         uint64 `field:"attackBytes"`         // 攻击流量
}

type TrafficDailyStatOperator struct {
	Id                  interface{} // ID
	Day                 interface{} // YYYYMMDD
	CachedBytes         interface{} // 缓存流量
	Bytes               interface{} // 流量字节
	CountRequests       interface{} // 请求数
	CountCachedRequests interface{} // 缓存请求数
	CountAttackRequests interface{} // 攻击量
	AttackBytes         interface{} // 攻击流量
}

func NewTrafficDailyStatOperator() *TrafficDailyStatOperator {
	return &TrafficDailyStatOperator{}
}
