package regions

//
type RegionCountry struct {
	Id     uint32 `field:"id"`     // ID
	Name   string `field:"name"`   // 名称
	Codes  string `field:"codes"`  // 代号
	State  uint8  `field:"state"`  // 状态
	DataId string `field:"dataId"` // 原始数据ID
	Pinyin string `field:"pinyin"` // 拼音
}

type RegionCountryOperator struct {
	Id     interface{} // ID
	Name   interface{} // 名称
	Codes  interface{} // 代号
	State  interface{} // 状态
	DataId interface{} // 原始数据ID
	Pinyin interface{} // 拼音
}

func NewRegionCountryOperator() *RegionCountryOperator {
	return &RegionCountryOperator{}
}
