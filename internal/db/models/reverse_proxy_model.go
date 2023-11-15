package models

import "github.com/iwind/TeaGo/dbs"

const (
	ReverseProxyField_Id                       dbs.FieldName = "id"                       // ID
	ReverseProxyField_AdminId                  dbs.FieldName = "adminId"                  // 管理员ID
	ReverseProxyField_UserId                   dbs.FieldName = "userId"                   // 用户ID
	ReverseProxyField_TemplateId               dbs.FieldName = "templateId"               // 模版ID
	ReverseProxyField_IsOn                     dbs.FieldName = "isOn"                     // 是否启用
	ReverseProxyField_Scheduling               dbs.FieldName = "scheduling"               // 调度算法
	ReverseProxyField_PrimaryOrigins           dbs.FieldName = "primaryOrigins"           // 主要源站
	ReverseProxyField_BackupOrigins            dbs.FieldName = "backupOrigins"            // 备用源站
	ReverseProxyField_StripPrefix              dbs.FieldName = "stripPrefix"              // 去除URL前缀
	ReverseProxyField_RequestHostType          dbs.FieldName = "requestHostType"          // 请求Host类型
	ReverseProxyField_RequestHost              dbs.FieldName = "requestHost"              // 请求Host
	ReverseProxyField_RequestHostExcludingPort dbs.FieldName = "requestHostExcludingPort" // 移除请求Host中的域名
	ReverseProxyField_RequestURI               dbs.FieldName = "requestURI"               // 请求URI
	ReverseProxyField_AutoFlush                dbs.FieldName = "autoFlush"                // 是否自动刷新缓冲区
	ReverseProxyField_AddHeaders               dbs.FieldName = "addHeaders"               // 自动添加的Header列表
	ReverseProxyField_State                    dbs.FieldName = "state"                    // 状态
	ReverseProxyField_CreatedAt                dbs.FieldName = "createdAt"                // 创建时间
	ReverseProxyField_ConnTimeout              dbs.FieldName = "connTimeout"              // 连接超时时间
	ReverseProxyField_ReadTimeout              dbs.FieldName = "readTimeout"              // 读取超时时间
	ReverseProxyField_IdleTimeout              dbs.FieldName = "idleTimeout"              // 空闲超时时间
	ReverseProxyField_MaxConns                 dbs.FieldName = "maxConns"                 // 最大并发连接数
	ReverseProxyField_MaxIdleConns             dbs.FieldName = "maxIdleConns"             // 最大空闲连接数
	ReverseProxyField_ProxyProtocol            dbs.FieldName = "proxyProtocol"            // Proxy Protocol配置
	ReverseProxyField_FollowRedirects          dbs.FieldName = "followRedirects"          // 回源跟随
	ReverseProxyField_Retry50X                 dbs.FieldName = "retry50X"                 // 启用50X重试
	ReverseProxyField_Retry40X                 dbs.FieldName = "retry40X"                 // 启用40X重试
)

// ReverseProxy 反向代理配置
type ReverseProxy struct {
	Id                       uint32   `field:"id"`                       // ID
	AdminId                  uint32   `field:"adminId"`                  // 管理员ID
	UserId                   uint32   `field:"userId"`                   // 用户ID
	TemplateId               uint32   `field:"templateId"`               // 模版ID
	IsOn                     bool     `field:"isOn"`                     // 是否启用
	Scheduling               dbs.JSON `field:"scheduling"`               // 调度算法
	PrimaryOrigins           dbs.JSON `field:"primaryOrigins"`           // 主要源站
	BackupOrigins            dbs.JSON `field:"backupOrigins"`            // 备用源站
	StripPrefix              string   `field:"stripPrefix"`              // 去除URL前缀
	RequestHostType          uint8    `field:"requestHostType"`          // 请求Host类型
	RequestHost              string   `field:"requestHost"`              // 请求Host
	RequestHostExcludingPort bool     `field:"requestHostExcludingPort"` // 移除请求Host中的域名
	RequestURI               string   `field:"requestURI"`               // 请求URI
	AutoFlush                uint8    `field:"autoFlush"`                // 是否自动刷新缓冲区
	AddHeaders               dbs.JSON `field:"addHeaders"`               // 自动添加的Header列表
	State                    uint8    `field:"state"`                    // 状态
	CreatedAt                uint64   `field:"createdAt"`                // 创建时间
	ConnTimeout              dbs.JSON `field:"connTimeout"`              // 连接超时时间
	ReadTimeout              dbs.JSON `field:"readTimeout"`              // 读取超时时间
	IdleTimeout              dbs.JSON `field:"idleTimeout"`              // 空闲超时时间
	MaxConns                 uint32   `field:"maxConns"`                 // 最大并发连接数
	MaxIdleConns             uint32   `field:"maxIdleConns"`             // 最大空闲连接数
	ProxyProtocol            dbs.JSON `field:"proxyProtocol"`            // Proxy Protocol配置
	FollowRedirects          uint8    `field:"followRedirects"`          // 回源跟随
	Retry50X                 bool     `field:"retry50X"`                 // 启用50X重试
	Retry40X                 bool     `field:"retry40X"`                 // 启用40X重试
}

type ReverseProxyOperator struct {
	Id                       any // ID
	AdminId                  any // 管理员ID
	UserId                   any // 用户ID
	TemplateId               any // 模版ID
	IsOn                     any // 是否启用
	Scheduling               any // 调度算法
	PrimaryOrigins           any // 主要源站
	BackupOrigins            any // 备用源站
	StripPrefix              any // 去除URL前缀
	RequestHostType          any // 请求Host类型
	RequestHost              any // 请求Host
	RequestHostExcludingPort any // 移除请求Host中的域名
	RequestURI               any // 请求URI
	AutoFlush                any // 是否自动刷新缓冲区
	AddHeaders               any // 自动添加的Header列表
	State                    any // 状态
	CreatedAt                any // 创建时间
	ConnTimeout              any // 连接超时时间
	ReadTimeout              any // 读取超时时间
	IdleTimeout              any // 空闲超时时间
	MaxConns                 any // 最大并发连接数
	MaxIdleConns             any // 最大空闲连接数
	ProxyProtocol            any // Proxy Protocol配置
	FollowRedirects          any // 回源跟随
	Retry50X                 any // 启用50X重试
	Retry40X                 any // 启用40X重试
}

func NewReverseProxyOperator() *ReverseProxyOperator {
	return &ReverseProxyOperator{}
}
