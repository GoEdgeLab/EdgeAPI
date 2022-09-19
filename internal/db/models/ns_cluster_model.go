package models

import "github.com/iwind/TeaGo/dbs"

// NSCluster 域名服务器集群
type NSCluster struct {
	Id              uint32   `field:"id"`              // ID
	IsOn            bool     `field:"isOn"`            // 是否启用
	Name            string   `field:"name"`            // 集群名
	InstallDir      string   `field:"installDir"`      // 安装目录
	State           uint8    `field:"state"`           // 状态
	AccessLog       dbs.JSON `field:"accessLog"`       // 访问日志配置
	GrantId         uint32   `field:"grantId"`         // 授权ID
	Recursion       dbs.JSON `field:"recursion"`       // 递归DNS设置
	Tcp             dbs.JSON `field:"tcp"`             // TCP设置
	Tls             dbs.JSON `field:"tls"`             // TLS设置
	Udp             dbs.JSON `field:"udp"`             // UDP设置
	DdosProtection  dbs.JSON `field:"ddosProtection"`  // DDoS防护设置
	Hosts           dbs.JSON `field:"hosts"`           // DNS主机地址
	AutoRemoteStart bool     `field:"autoRemoteStart"` // 自动远程启动
	TimeZone        string   `field:"timeZone"`        // 时区
}

type NSClusterOperator struct {
	Id              any // ID
	IsOn            any // 是否启用
	Name            any // 集群名
	InstallDir      any // 安装目录
	State           any // 状态
	AccessLog       any // 访问日志配置
	GrantId         any // 授权ID
	Recursion       any // 递归DNS设置
	Tcp             any // TCP设置
	Tls             any // TLS设置
	Udp             any // UDP设置
	DdosProtection  any // DDoS防护设置
	Hosts           any // DNS主机地址
	AutoRemoteStart any // 自动远程启动
	TimeZone        any // 时区
}

func NewNSClusterOperator() *NSClusterOperator {
	return &NSClusterOperator{}
}
