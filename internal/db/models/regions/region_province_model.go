package regions

import "github.com/iwind/TeaGo/dbs"

const (
	RegionProvinceField_Id          dbs.FieldName = "id"          // ID
	RegionProvinceField_ValueId     dbs.FieldName = "valueId"     // 实际ID
	RegionProvinceField_CountryId   dbs.FieldName = "countryId"   // 国家ID
	RegionProvinceField_Name        dbs.FieldName = "name"        // 名称
	RegionProvinceField_Codes       dbs.FieldName = "codes"       // 代号
	RegionProvinceField_CustomName  dbs.FieldName = "customName"  // 自定义名称
	RegionProvinceField_CustomCodes dbs.FieldName = "customCodes" // 自定义代号
	RegionProvinceField_State       dbs.FieldName = "state"       // 状态
	RegionProvinceField_DataId      dbs.FieldName = "dataId"      // 原始数据ID
	RegionProvinceField_RouteCode   dbs.FieldName = "routeCode"   // 线路代号
)

// RegionProvince 区域-省份
type RegionProvince struct {
	Id          uint32   `field:"id"`          // ID
	ValueId     uint32   `field:"valueId"`     // 实际ID
	CountryId   uint32   `field:"countryId"`   // 国家ID
	Name        string   `field:"name"`        // 名称
	Codes       dbs.JSON `field:"codes"`       // 代号
	CustomName  string   `field:"customName"`  // 自定义名称
	CustomCodes dbs.JSON `field:"customCodes"` // 自定义代号
	State       uint8    `field:"state"`       // 状态
	DataId      string   `field:"dataId"`      // 原始数据ID
	RouteCode   string   `field:"routeCode"`   // 线路代号
}

type RegionProvinceOperator struct {
	Id          any // ID
	ValueId     any // 实际ID
	CountryId   any // 国家ID
	Name        any // 名称
	Codes       any // 代号
	CustomName  any // 自定义名称
	CustomCodes any // 自定义代号
	State       any // 状态
	DataId      any // 原始数据ID
	RouteCode   any // 线路代号
}

func NewRegionProvinceOperator() *RegionProvinceOperator {
	return &RegionProvinceOperator{}
}
