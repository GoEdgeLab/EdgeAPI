package regions

import "github.com/iwind/TeaGo/dbs"

// RegionProvider 区域-运营商
type RegionProvider struct {
	Id          uint32   `field:"id"`          // ID
	Name        string   `field:"name"`        // 名称
	Codes       dbs.JSON `field:"codes"`       // 代号
	CustomName  string   `field:"customName"`  // 自定义名称
	CustomCodes dbs.JSON `field:"customCodes"` // 自定义代号
	State       uint8    `field:"state"`       // 状态
}

type RegionProviderOperator struct {
	Id          interface{} // ID
	Name        interface{} // 名称
	Codes       interface{} // 代号
	CustomName  interface{} // 自定义名称
	CustomCodes interface{} // 自定义代号
	State       interface{} // 状态
}

func NewRegionProviderOperator() *RegionProviderOperator {
	return &RegionProviderOperator{}
}
