package stats

import "github.com/iwind/TeaGo/dbs"

const (
	TrafficDailyStatField_Id                  dbs.FieldName = "id"                  // ID
	TrafficDailyStatField_Day                 dbs.FieldName = "day"                 // YYYYMMDD
	TrafficDailyStatField_CachedBytes         dbs.FieldName = "cachedBytes"         // 缓存流量
	TrafficDailyStatField_Bytes               dbs.FieldName = "bytes"               // 流量字节
	TrafficDailyStatField_CountRequests       dbs.FieldName = "countRequests"       // 请求数
	TrafficDailyStatField_CountCachedRequests dbs.FieldName = "countCachedRequests" // 缓存请求数
	TrafficDailyStatField_CountAttackRequests dbs.FieldName = "countAttackRequests" // 攻击量
	TrafficDailyStatField_AttackBytes         dbs.FieldName = "attackBytes"         // 攻击流量
	TrafficDailyStatField_CountIPs            dbs.FieldName = "countIPs"            // 独立IP数
)

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
	CountIPs            uint64 `field:"countIPs"`            // 独立IP数
}

type TrafficDailyStatOperator struct {
	Id                  any // ID
	Day                 any // YYYYMMDD
	CachedBytes         any // 缓存流量
	Bytes               any // 流量字节
	CountRequests       any // 请求数
	CountCachedRequests any // 缓存请求数
	CountAttackRequests any // 攻击量
	AttackBytes         any // 攻击流量
	CountIPs            any // 独立IP数
}

func NewTrafficDailyStatOperator() *TrafficDailyStatOperator {
	return &TrafficDailyStatOperator{}
}
