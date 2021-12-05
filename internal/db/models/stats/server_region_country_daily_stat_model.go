package stats

// ServerRegionCountryDailyStat 服务用户区域分布统计（按天）
type ServerRegionCountryDailyStat struct {
	Id                  uint64 `field:"id"`                  // ID
	ServerId            uint32 `field:"serverId"`            // 服务ID
	CountryId           uint32 `field:"countryId"`           // 国家/区域ID
	Day                 string `field:"day"`                 // 日期YYYYMMDD
	CountRequests       uint64 `field:"countRequests"`       // 请求数量
	CountAttackRequests uint64 `field:"countAttackRequests"` // 攻击数量
	AttackBytes         uint64 `field:"attackBytes"`         // 攻击流量
	Bytes               uint64 `field:"bytes"`               // 总流量
}

type ServerRegionCountryDailyStatOperator struct {
	Id                  interface{} // ID
	ServerId            interface{} // 服务ID
	CountryId           interface{} // 国家/区域ID
	Day                 interface{} // 日期YYYYMMDD
	CountRequests       interface{} // 请求数量
	CountAttackRequests interface{} // 攻击数量
	AttackBytes         interface{} // 攻击流量
	Bytes               interface{} // 总流量
}

func NewServerRegionCountryDailyStatOperator() *ServerRegionCountryDailyStatOperator {
	return &ServerRegionCountryDailyStatOperator{}
}
