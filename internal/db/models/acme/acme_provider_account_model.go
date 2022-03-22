package acme

// ACMEProviderAccount ACME提供商
type ACMEProviderAccount struct {
	Id           uint64 `field:"id"`           // ID
	IsOn         bool   `field:"isOn"`         // 是否启用
	Name         string `field:"name"`         // 名称
	ProviderCode string `field:"providerCode"` // 代号
	Error        string `field:"error"`        // 最后一条错误信息
	EabKid       string `field:"eabKid"`       // KID
	EabKey       string `field:"eabKey"`       // Key
	State        uint8  `field:"state"`        // 状态
}

type ACMEProviderAccountOperator struct {
	Id           interface{} // ID
	IsOn         interface{} // 是否启用
	Name         interface{} // 名称
	ProviderCode interface{} // 代号
	Error        interface{} // 最后一条错误信息
	EabKid       interface{} // KID
	EabKey       interface{} // Key
	State        interface{} // 状态
}

func NewACMEProviderAccountOperator() *ACMEProviderAccountOperator {
	return &ACMEProviderAccountOperator{}
}
