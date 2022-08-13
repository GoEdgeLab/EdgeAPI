package regions

import "github.com/iwind/TeaGo/dbs"

// RegionCountry 区域-国家/地区
type RegionCountry struct {
	Id          uint32   `field:"id"`          // ID
	Name        string   `field:"name"`        // 名称
	Codes       dbs.JSON `field:"codes"`       // 代号
	CustomName  string   `field:"customName"`  // 自定义名称
	CustomCodes dbs.JSON `field:"customCodes"` // 自定义代号
	State       uint8    `field:"state"`       // 状态
	DataId      string   `field:"dataId"`      // 原始数据ID
	Pinyin      dbs.JSON `field:"pinyin"`      // 拼音
}

type RegionCountryOperator struct {
	Id          interface{} // ID
	Name        interface{} // 名称
	Codes       interface{} // 代号
	CustomName  interface{} // 自定义名称
	CustomCodes interface{} // 自定义代号
	State       interface{} // 状态
	DataId      interface{} // 原始数据ID
	Pinyin      interface{} // 拼音
}

func NewRegionCountryOperator() *RegionCountryOperator {
	return &RegionCountryOperator{}
}
