package acme

// ACMEProviderAccount ACME提供商
type ACMEProviderAccount struct {
	Id           uint64 `field:"id"`           // ID
	UserId       uint64 `field:"userId"`       // 用户ID
	IsOn         bool   `field:"isOn"`         // 是否启用
	Name         string `field:"name"`         // 名称
	ProviderCode string `field:"providerCode"` // 代号
	EabKid       string `field:"eabKid"`       // KID
	EabKey       string `field:"eabKey"`       // Key
	Error        string `field:"error"`        // 最后一条错误信息
	State        uint8  `field:"state"`        // 状态
}

type ACMEProviderAccountOperator struct {
	Id           any // ID
	UserId       any // 用户ID
	IsOn         any // 是否启用
	Name         any // 名称
	ProviderCode any // 代号
	EabKid       any // KID
	EabKey       any // Key
	Error        any // 最后一条错误信息
	State        any // 状态
}

func NewACMEProviderAccountOperator() *ACMEProviderAccountOperator {
	return &ACMEProviderAccountOperator{}
}
