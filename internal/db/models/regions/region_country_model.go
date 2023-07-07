package regions

import "github.com/iwind/TeaGo/dbs"

const (
	RegionCountryField_Id          dbs.FieldName = "id"          // ID
	RegionCountryField_ValueId     dbs.FieldName = "valueId"     // 实际ID
	RegionCountryField_ValueCode   dbs.FieldName = "valueCode"   // 值代号
	RegionCountryField_Name        dbs.FieldName = "name"        // 名称
	RegionCountryField_Codes       dbs.FieldName = "codes"       // 代号
	RegionCountryField_CustomName  dbs.FieldName = "customName"  // 自定义名称
	RegionCountryField_CustomCodes dbs.FieldName = "customCodes" // 自定义代号
	RegionCountryField_State       dbs.FieldName = "state"       // 状态
	RegionCountryField_DataId      dbs.FieldName = "dataId"      // 原始数据ID
	RegionCountryField_Pinyin      dbs.FieldName = "pinyin"      // 拼音
	RegionCountryField_IsCommon    dbs.FieldName = "isCommon"    // 是否常用
)

// RegionCountry 区域-国家/地区
type RegionCountry struct {
	Id1          uint32   `field:"id"`          // ID
	ValueId     uint32   `field:"valueId"`     // 实际ID
	ValueCode   string   `field:"valueCode"`   // 值代号
	Name        string   `field:"name"`        // 名称
	Codes       dbs.JSON `field:"codes"`       // 代号
	CustomName  string   `field:"customName"`  // 自定义名称
	CustomCodes dbs.JSON `field:"customCodes"` // 自定义代号
	State       uint8    `field:"state"`       // 状态
	DataId      string   `field:"dataId"`      // 原始数据ID
	Pinyin      dbs.JSON `field:"pinyin"`      // 拼音
	IsCommon    bool     `field:"isCommon"`    // 是否常用
}

type RegionCountryOperator struct {
	Id          any // ID
	ValueId     any // 实际ID
	ValueCode   any // 值代号
	Name        any // 名称
	Codes       any // 代号
	CustomName  any // 自定义名称
	CustomCodes any // 自定义代号
	State       any // 状态
	DataId      any // 原始数据ID
	Pinyin      any // 拼音
	IsCommon    any // 是否常用
}

func NewRegionCountryOperator() *RegionCountryOperator {
	return &RegionCountryOperator{}
}
