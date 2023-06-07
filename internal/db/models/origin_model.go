package models

import "github.com/iwind/TeaGo/dbs"

// Origin 源站
type Origin struct {
	Id                 uint32   `field:"id"`                 // ID
	AdminId            uint32   `field:"adminId"`            // 管理员ID
	UserId             uint32   `field:"userId"`             // 用户ID
	IsOn               bool     `field:"isOn"`               // 是否启用
	Name               string   `field:"name"`               // 名称
	Version            uint32   `field:"version"`            // 版本
	Addr               dbs.JSON `field:"addr"`               // 地址
	Oss                dbs.JSON `field:"oss"`                // OSS配置
	Description        string   `field:"description"`        // 描述
	Code               string   `field:"code"`               // 代号
	Weight             uint32   `field:"weight"`             // 权重
	ConnTimeout        dbs.JSON `field:"connTimeout"`        // 连接超时
	ReadTimeout        dbs.JSON `field:"readTimeout"`        // 读超时
	IdleTimeout        dbs.JSON `field:"idleTimeout"`        // 空闲连接超时
	MaxFails           uint32   `field:"maxFails"`           // 最多失败次数
	MaxConns           uint32   `field:"maxConns"`           // 最大并发连接数
	MaxIdleConns       uint32   `field:"maxIdleConns"`       // 最多空闲连接数
	HttpRequestURI     string   `field:"httpRequestURI"`     // 转发后的请求URI
	HttpRequestHeader  dbs.JSON `field:"httpRequestHeader"`  // 请求Header配置
	HttpResponseHeader dbs.JSON `field:"httpResponseHeader"` // 响应Header配置
	Host               string   `field:"host"`               // 自定义主机名
	HealthCheck        dbs.JSON `field:"healthCheck"`        // 健康检查设置
	Cert               dbs.JSON `field:"cert"`               // 证书设置
	Ftp                dbs.JSON `field:"ftp"`                // FTP相关设置
	CreatedAt          uint64   `field:"createdAt"`          // 创建时间
	Domains            dbs.JSON `field:"domains"`            // 所属域名
	FollowPort         bool     `field:"followPort"`         // 端口跟随
	State              uint8    `field:"state"`              // 状态
}

type OriginOperator struct {
	Id                 any // ID
	AdminId            any // 管理员ID
	UserId             any // 用户ID
	IsOn               any // 是否启用
	Name               any // 名称
	Version            any // 版本
	Addr               any // 地址
	Oss                any // OSS配置
	Description        any // 描述
	Code               any // 代号
	Weight             any // 权重
	ConnTimeout        any // 连接超时
	ReadTimeout        any // 读超时
	IdleTimeout        any // 空闲连接超时
	MaxFails           any // 最多失败次数
	MaxConns           any // 最大并发连接数
	MaxIdleConns       any // 最多空闲连接数
	HttpRequestURI     any // 转发后的请求URI
	HttpRequestHeader  any // 请求Header配置
	HttpResponseHeader any // 响应Header配置
	Host               any // 自定义主机名
	HealthCheck        any // 健康检查设置
	Cert               any // 证书设置
	Ftp                any // FTP相关设置
	CreatedAt          any // 创建时间
	Domains            any // 所属域名
	FollowPort         any // 端口跟随
	State              any // 状态
}

func NewOriginOperator() *OriginOperator {
	return &OriginOperator{}
}
