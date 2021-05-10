package models

// HTTPFastcgi Fastcgi设置
type HTTPFastcgi struct {
	Id              uint64 `field:"id"`              // ID
	AdminId         uint32 `field:"adminId"`         // 管理员ID
	UserId          uint32 `field:"userId"`          // 用户ID
	IsOn            uint8  `field:"isOn"`            // 是否启用
	Address         string `field:"address"`         // 地址
	Params          string `field:"params"`          // 参数
	ReadTimeout     string `field:"readTimeout"`     // 读取超时
	ConnTimeout     string `field:"connTimeout"`     // 连接超时
	PoolSize        uint32 `field:"poolSize"`        // 连接池尺寸
	PathInfoPattern string `field:"pathInfoPattern"` // PATH_INFO匹配
	State           uint8  `field:"state"`           // 状态
}

type HTTPFastcgiOperator struct {
	Id              interface{} // ID
	AdminId         interface{} // 管理员ID
	UserId          interface{} // 用户ID
	IsOn            interface{} // 是否启用
	Address         interface{} // 地址
	Params          interface{} // 参数
	ReadTimeout     interface{} // 读取超时
	ConnTimeout     interface{} // 连接超时
	PoolSize        interface{} // 连接池尺寸
	PathInfoPattern interface{} // PATH_INFO匹配
	State           interface{} // 状态
}

func NewHTTPFastcgiOperator() *HTTPFastcgiOperator {
	return &HTTPFastcgiOperator{}
}
