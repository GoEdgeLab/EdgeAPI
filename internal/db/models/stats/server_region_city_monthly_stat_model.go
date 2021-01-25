package stats

// 服务用户省份分布统计（按天）
type ServerRegionCityMonthlyStat struct {
	Id       uint64 `field:"id"`       // ID
	ServerId uint32 `field:"serverId"` // 服务ID
	CityId   uint32 `field:"cityId"`   // 城市ID
	Month    string `field:"month"`    // 月份YYYYMM
	Count    uint64 `field:"count"`    // 数量
}

type ServerRegionCityMonthlyStatOperator struct {
	Id       interface{} // ID
	ServerId interface{} // 服务ID
	CityId   interface{} // 城市ID
	Month    interface{} // 月份YYYYMM
	Count    interface{} // 数量
}

func NewServerRegionCityMonthlyStatOperator() *ServerRegionCityMonthlyStatOperator {
	return &ServerRegionCityMonthlyStatOperator{}
}
