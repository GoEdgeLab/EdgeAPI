package regions

import "github.com/iwind/TeaGo/dbs"

// RegionTown 区域-省份
type RegionTown struct {
	Id          uint32   `field:"id"`          // ID
	CityId      uint32   `field:"cityId"`      // 城市ID
	Name        string   `field:"name"`        // 名称
	Codes       dbs.JSON `field:"codes"`       // 代号
	CustomName  string   `field:"customName"`  // 自定义名称
	CustomCodes dbs.JSON `field:"customCodes"` // 自定义代号
	State       uint8    `field:"state"`       // 状态
	DataId      string   `field:"dataId"`      // 原始数据ID
}

type RegionTownOperator struct {
	Id          interface{} // ID
	CityId      interface{} // 城市ID
	Name        interface{} // 名称
	Codes       interface{} // 代号
	CustomName  interface{} // 自定义名称
	CustomCodes interface{} // 自定义代号
	State       interface{} // 状态
	DataId      interface{} // 原始数据ID
}

func NewRegionTownOperator() *RegionTownOperator {
	return &RegionTownOperator{}
}
