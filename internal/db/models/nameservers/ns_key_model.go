package nameservers

// NSKey 密钥管理
type NSKey struct {
	Id         uint64 `field:"id"`         // ID
	IsOn       uint8  `field:"isOn"`       // 状态
	Name       string `field:"name"`       // 名称
	DomainId   uint64 `field:"domainId"`   // 域名ID
	ZoneId     uint64 `field:"zoneId"`     // 子域ID
	Algo       string `field:"algo"`       // 算法
	Secret     string `field:"secret"`     // 密码
	SecretType string `field:"secretType"` // 密码类型
	Version    uint64 `field:"version"`    // 版本号
	State      uint8  `field:"state"`      // 状态
}

type NSKeyOperator struct {
	Id         interface{} // ID
	IsOn       interface{} // 状态
	Name       interface{} // 名称
	DomainId   interface{} // 域名ID
	ZoneId     interface{} // 子域ID
	Algo       interface{} // 算法
	Secret     interface{} // 密码
	SecretType interface{} // 密码类型
	Version    interface{} // 版本号
	State      interface{} // 状态
}

func NewNSKeyOperator() *NSKeyOperator {
	return &NSKeyOperator{}
}
