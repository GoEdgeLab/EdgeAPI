package stats

// 服务用户区域分布统计（按天）
type ServerRegionCountryMonthlyStat struct {
	Id        uint64 `field:"id"`        // ID
	ServerId  uint32 `field:"serverId"`  // 服务ID
	CountryId uint32 `field:"countryId"` // 国家/区域ID
	Month     string `field:"month"`     // 月份YYYYMM
	Count     uint64 `field:"count"`     // 数量
}

type ServerRegionCountryMonthlyStatOperator struct {
	Id        interface{} // ID
	ServerId  interface{} // 服务ID
	CountryId interface{} // 国家/区域ID
	Month     interface{} // 月份YYYYMM
	Count     interface{} // 数量
}

func NewServerRegionCountryMonthlyStatOperator() *ServerRegionCountryMonthlyStatOperator {
	return &ServerRegionCountryMonthlyStatOperator{}
}
