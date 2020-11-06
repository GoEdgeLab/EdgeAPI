package models

//
type RegionProvince struct {
	Id        uint32 `field:"id"`        // ID
	CountryId uint32 `field:"countryId"` // 国家ID
	Name      string `field:"name"`      // 名称
	Codes     string `field:"codes"`     // 代号
	State     uint8  `field:"state"`     // 状态
	DataId    string `field:"dataId"`    // 原始数据ID
}

type RegionProvinceOperator struct {
	Id        interface{} // ID
	CountryId interface{} // 国家ID
	Name      interface{} // 名称
	Codes     interface{} // 代号
	State     interface{} // 状态
	DataId    interface{} // 原始数据ID
}

func NewRegionProvinceOperator() *RegionProvinceOperator {
	return &RegionProvinceOperator{}
}
