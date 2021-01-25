package regions

//
type RegionProvider struct {
	Id    uint32 `field:"id"`    // ID
	Name  string `field:"name"`  // 名称
	Codes string `field:"codes"` // 代号
	State uint8  `field:"state"` // 状态
}

type RegionProviderOperator struct {
	Id    interface{} // ID
	Name  interface{} // 名称
	Codes interface{} // 代号
	State interface{} // 状态
}

func NewRegionProviderOperator() *RegionProviderOperator {
	return &RegionProviderOperator{}
}
