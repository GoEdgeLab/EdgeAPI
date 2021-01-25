package stats

// 服务用户省份分布统计（按天）
type ServerRegionProviderMonthlyStat struct {
	Id         uint64 `field:"id"`         // ID
	ServerId   uint32 `field:"serverId"`   // 服务ID
	ProviderId uint32 `field:"providerId"` // 运营商ID
	Month      string `field:"month"`      // 月份YYYYMM
	Count      uint64 `field:"count"`      // 数量
}

type ServerRegionProviderMonthlyStatOperator struct {
	Id         interface{} // ID
	ServerId   interface{} // 服务ID
	ProviderId interface{} // 运营商ID
	Month      interface{} // 月份YYYYMM
	Count      interface{} // 数量
}

func NewServerRegionProviderMonthlyStatOperator() *ServerRegionProviderMonthlyStatOperator {
	return &ServerRegionProviderMonthlyStatOperator{}
}
