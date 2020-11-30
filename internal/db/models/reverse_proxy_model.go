package models

// 反向代理配置
type ReverseProxy struct {
	Id              uint32 `field:"id"`              // ID
	AdminId         uint32 `field:"adminId"`         // 管理员ID
	UserId          uint32 `field:"userId"`          // 用户ID
	TemplateId      uint32 `field:"templateId"`      // 模版ID
	IsOn            uint8  `field:"isOn"`            // 是否启用
	Scheduling      string `field:"scheduling"`      // 调度算法
	PrimaryOrigins  string `field:"primaryOrigins"`  // 主要源站
	BackupOrigins   string `field:"backupOrigins"`   // 备用源站
	StripPrefix     string `field:"stripPrefix"`     // 去除URL前缀
	RequestHostType uint8  `field:"requestHostType"` // 请求Host类型
	RequestHost     string `field:"requestHost"`     // 请求Host
	RequestURI      string `field:"requestURI"`      // 请求URI
	AutoFlush       uint8  `field:"autoFlush"`       // 是否自动刷新缓冲区
	State           uint8  `field:"state"`           // 状态
	CreatedAt       uint64 `field:"createdAt"`       // 创建时间
}

type ReverseProxyOperator struct {
	Id              interface{} // ID
	AdminId         interface{} // 管理员ID
	UserId          interface{} // 用户ID
	TemplateId      interface{} // 模版ID
	IsOn            interface{} // 是否启用
	Scheduling      interface{} // 调度算法
	PrimaryOrigins  interface{} // 主要源站
	BackupOrigins   interface{} // 备用源站
	StripPrefix     interface{} // 去除URL前缀
	RequestHostType interface{} // 请求Host类型
	RequestHost     interface{} // 请求Host
	RequestURI      interface{} // 请求URI
	AutoFlush       interface{} // 是否自动刷新缓冲区
	State           interface{} // 状态
	CreatedAt       interface{} // 创建时间
}

func NewReverseProxyOperator() *ReverseProxyOperator {
	return &ReverseProxyOperator{}
}
