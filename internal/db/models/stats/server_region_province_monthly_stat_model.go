package stats

// 服务用户省份分布统计（按天）
type ServerRegionProvinceMonthlyStat struct {
	Id         uint64 `field:"id"`         // ID
	ServerId   uint32 `field:"serverId"`   // 服务ID
	ProvinceId uint32 `field:"provinceId"` // 省份ID
	Month      string `field:"month"`      // 月份YYYYMM
	Count      uint64 `field:"count"`      // 数量
}

type ServerRegionProvinceMonthlyStatOperator struct {
	Id         interface{} // ID
	ServerId   interface{} // 服务ID
	ProvinceId interface{} // 省份ID
	Month      interface{} // 月份YYYYMM
	Count      interface{} // 数量
}

func NewServerRegionProvinceMonthlyStatOperator() *ServerRegionProvinceMonthlyStatOperator {
	return &ServerRegionProvinceMonthlyStatOperator{}
}
