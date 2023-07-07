package regions

import "github.com/iwind/TeaGo/dbs"

const (
	RegionTownField_Id          dbs.FieldName = "id"          // ID
	RegionTownField_ValueId     dbs.FieldName = "valueId"     // 真实ID
	RegionTownField_CityId      dbs.FieldName = "cityId"      // 城市ID
	RegionTownField_Name        dbs.FieldName = "name"        // 名称
	RegionTownField_Codes       dbs.FieldName = "codes"       // 代号
	RegionTownField_CustomName  dbs.FieldName = "customName"  // 自定义名称
	RegionTownField_CustomCodes dbs.FieldName = "customCodes" // 自定义代号
	RegionTownField_State       dbs.FieldName = "state"       // 状态
	RegionTownField_DataId      dbs.FieldName = "dataId"      // 原始数据ID
)

// RegionTown 区域-省份
type RegionTown struct {
	Id1          uint32   `field:"id"`          // ID
	ValueId     uint32   `field:"valueId"`     // 真实ID
	CityId      uint32   `field:"cityId"`      // 城市ID
	Name        string   `field:"name"`        // 名称
	Codes       dbs.JSON `field:"codes"`       // 代号
	CustomName  string   `field:"customName"`  // 自定义名称
	CustomCodes dbs.JSON `field:"customCodes"` // 自定义代号
	State       uint8    `field:"state"`       // 状态
	DataId      string   `field:"dataId"`      // 原始数据ID
}

type RegionTownOperator struct {
	Id          any // ID
	ValueId     any // 真实ID
	CityId      any // 城市ID
	Name        any // 名称
	Codes       any // 代号
	CustomName  any // 自定义名称
	CustomCodes any // 自定义代号
	State       any // 状态
	DataId      any // 原始数据ID
}

func NewRegionTownOperator() *RegionTownOperator {
	return &RegionTownOperator{}
}
