package models

// 源站
type Origin struct {
	Id                 uint32 `field:"id"`                 // ID
	AdminId            uint32 `field:"adminId"`            // 管理员ID
	UserId             uint32 `field:"userId"`             // 用户ID
	IsOn               uint8  `field:"isOn"`               // 是否启用
	Name               string `field:"name"`               // 名称
	Version            uint32 `field:"version"`            // 版本
	Addr               string `field:"addr"`               // 地址
	Description        string `field:"description"`        // 描述
	Code               string `field:"code"`               // 代号
	Weight             uint32 `field:"weight"`             // 权重
	ConnTimeout        string `field:"connTimeout"`        // 连接超时
	ReadTimeout        string `field:"readTimeout"`        // 读超时
	IdleTimeout        string `field:"idleTimeout"`        // 空闲连接超时
	MaxFails           uint32 `field:"maxFails"`           // 最多失败次数
	MaxConns           uint32 `field:"maxConns"`           // 最大并发连接数
	MaxIdleConns       uint32 `field:"maxIdleConns"`       // 最多空闲连接数
	HttpRequestURI     string `field:"httpRequestURI"`     // 转发后的请求URI
	HttpRequestHeader  string `field:"httpRequestHeader"`  // 请求Header配置
	HttpResponseHeader string `field:"httpResponseHeader"` // 响应Header配置
	Host               string `field:"host"`               // 自定义主机名
	HealthCheck        string `field:"healthCheck"`        // 健康检查设置
	Cert               string `field:"cert"`               // 证书设置
	Ftp                string `field:"ftp"`                // FTP相关设置
	CreatedAt          uint64 `field:"createdAt"`          // 创建时间
	State              uint8  `field:"state"`              // 状态
}

type OriginOperator struct {
	Id                 interface{} // ID
	AdminId            interface{} // 管理员ID
	UserId             interface{} // 用户ID
	IsOn               interface{} // 是否启用
	Name               interface{} // 名称
	Version            interface{} // 版本
	Addr               interface{} // 地址
	Description        interface{} // 描述
	Code               interface{} // 代号
	Weight             interface{} // 权重
	ConnTimeout        interface{} // 连接超时
	ReadTimeout        interface{} // 读超时
	IdleTimeout        interface{} // 空闲连接超时
	MaxFails           interface{} // 最多失败次数
	MaxConns           interface{} // 最大并发连接数
	MaxIdleConns       interface{} // 最多空闲连接数
	HttpRequestURI     interface{} // 转发后的请求URI
	HttpRequestHeader  interface{} // 请求Header配置
	HttpResponseHeader interface{} // 响应Header配置
	Host               interface{} // 自定义主机名
	HealthCheck        interface{} // 健康检查设置
	Cert               interface{} // 证书设置
	Ftp                interface{} // FTP相关设置
	CreatedAt          interface{} // 创建时间
	State              interface{} // 状态
}

func NewOriginOperator() *OriginOperator {
	return &OriginOperator{}
}
