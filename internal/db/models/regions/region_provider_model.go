package regions

import "github.com/iwind/TeaGo/dbs"

const (
	RegionProviderField_Id          dbs.FieldName = "id"          // ID
	RegionProviderField_ValueId     dbs.FieldName = "valueId"     // 实际ID
	RegionProviderField_Name        dbs.FieldName = "name"        // 名称
	RegionProviderField_Codes       dbs.FieldName = "codes"       // 代号
	RegionProviderField_CustomName  dbs.FieldName = "customName"  // 自定义名称
	RegionProviderField_CustomCodes dbs.FieldName = "customCodes" // 自定义代号
	RegionProviderField_State       dbs.FieldName = "state"       // 状态
)

// RegionProvider 区域-运营商
type RegionProvider struct {
	Id1          uint32   `field:"id"`          // ID
	ValueId     uint32   `field:"valueId"`     // 实际ID
	Name        string   `field:"name"`        // 名称
	Codes       dbs.JSON `field:"codes"`       // 代号
	CustomName  string   `field:"customName"`  // 自定义名称
	CustomCodes dbs.JSON `field:"customCodes"` // 自定义代号
	State       uint8    `field:"state"`       // 状态
}

type RegionProviderOperator struct {
	Id          any // ID
	ValueId     any // 实际ID
	Name        any // 名称
	Codes       any // 代号
	CustomName  any // 自定义名称
	CustomCodes any // 自定义代号
	State       any // 状态
}

func NewRegionProviderOperator() *RegionProviderOperator {
	return &RegionProviderOperator{}
}
