package stats

// TrafficHourlyStat 总的流量统计（按小时）
type TrafficHourlyStat struct {
	Id                  uint64 `field:"id"`                  // ID
	Hour                string `field:"hour"`                // YYYYMMDDHH
	Bytes               uint64 `field:"bytes"`               // 流量字节
	CachedBytes         uint64 `field:"cachedBytes"`         // 缓存流量
	CountRequests       uint64 `field:"countRequests"`       // 请求数
	CountCachedRequests uint64 `field:"countCachedRequests"` // 缓存请求数
	CountAttackRequests uint64 `field:"countAttackRequests"` // 攻击数
	AttackBytes         uint64 `field:"attackBytes"`         // 攻击流量
}

type TrafficHourlyStatOperator struct {
	Id                  interface{} // ID
	Hour                interface{} // YYYYMMDDHH
	Bytes               interface{} // 流量字节
	CachedBytes         interface{} // 缓存流量
	CountRequests       interface{} // 请求数
	CountCachedRequests interface{} // 缓存请求数
	CountAttackRequests interface{} // 攻击数
	AttackBytes         interface{} // 攻击流量
}

func NewTrafficHourlyStatOperator() *TrafficHourlyStatOperator {
	return &TrafficHourlyStatOperator{}
}
