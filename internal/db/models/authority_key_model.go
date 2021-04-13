package models

// AuthorityKey 企业版认证信息
type AuthorityKey struct {
	Id           uint32 `field:"id"`           // ID
	Value        string `field:"value"`        // Key值
	DayFrom      string `field:"dayFrom"`      // 开始日期
	DayTo        string `field:"dayTo"`        // 结束日期
	Hostname     string `field:"hostname"`     // Hostname
	MacAddresses string `field:"macAddresses"` // MAC地址
	UpdatedAt    uint64 `field:"updatedAt"`    // 创建/修改时间
}

type AuthorityKeyOperator struct {
	Id           interface{} // ID
	Value        interface{} // Key值
	DayFrom      interface{} // 开始日期
	DayTo        interface{} // 结束日期
	Hostname     interface{} // Hostname
	MacAddresses interface{} // MAC地址
	UpdatedAt    interface{} // 创建/修改时间
}

func NewAuthorityKeyOperator() *AuthorityKeyOperator {
	return &AuthorityKeyOperator{}
}
