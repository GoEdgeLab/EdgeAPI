package regions

import "github.com/iwind/TeaGo/dbs"

const (
	RegionCityField_Id          dbs.FieldName = "id"          // ID
	RegionCityField_ValueId     dbs.FieldName = "valueId"     // 实际ID
	RegionCityField_ProvinceId  dbs.FieldName = "provinceId"  // 省份ID
	RegionCityField_Name        dbs.FieldName = "name"        // 名称
	RegionCityField_Codes       dbs.FieldName = "codes"       // 代号
	RegionCityField_CustomName  dbs.FieldName = "customName"  // 自定义名称
	RegionCityField_CustomCodes dbs.FieldName = "customCodes" // 自定义代号
	RegionCityField_State       dbs.FieldName = "state"       // 状态
	RegionCityField_DataId      dbs.FieldName = "dataId"      // 原始数据ID
)

// RegionCity 区域-城市
type RegionCity struct {
	Id1          uint32   `field:"id"`          // ID
	ValueId     uint32   `field:"valueId"`     // 实际ID
	ProvinceId  uint32   `field:"provinceId"`  // 省份ID
	Name        string   `field:"name"`        // 名称
	Codes       dbs.JSON `field:"codes"`       // 代号
	CustomName  string   `field:"customName"`  // 自定义名称
	CustomCodes dbs.JSON `field:"customCodes"` // 自定义代号
	State       uint8    `field:"state"`       // 状态
	DataId      string   `field:"dataId"`      // 原始数据ID
}

type RegionCityOperator struct {
	Id          any // ID
	ValueId     any // 实际ID
	ProvinceId  any // 省份ID
	Name        any // 名称
	Codes       any // 代号
	CustomName  any // 自定义名称
	CustomCodes any // 自定义代号
	State       any // 状态
	DataId      any // 原始数据ID
}

func NewRegionCityOperator() *RegionCityOperator {
	return &RegionCityOperator{}
}
