package models

import "github.com/iwind/TeaGo/dbs"

// NSNode 域名服务器节点
type NSNode struct {
	Id                 uint32   `field:"id"`                 // ID
	AdminId            uint32   `field:"adminId"`            // 管理员ID
	ClusterId          uint32   `field:"clusterId"`          // 集群ID
	Name               string   `field:"name"`               // 节点名称
	IsOn               bool     `field:"isOn"`               // 是否启用
	Status             dbs.JSON `field:"status"`             // 运行状态
	UniqueId           string   `field:"uniqueId"`           // 节点ID
	Secret             string   `field:"secret"`             // 密钥
	IsUp               bool     `field:"isUp"`               // 是否运行
	IsInstalled        bool     `field:"isInstalled"`        // 是否已安装
	InstallStatus      dbs.JSON `field:"installStatus"`      // 安装状态
	InstallDir         string   `field:"installDir"`         // 安装目录
	State              uint8    `field:"state"`              // 状态
	IsActive           bool     `field:"isActive"`           // 是否活跃
	StatusIsNotified   uint8    `field:"statusIsNotified"`   // 活跃状态已经通知
	InactiveNotifiedAt uint64   `field:"inactiveNotifiedAt"` // 离线通知时间
	ConnectedAPINodes  dbs.JSON `field:"connectedAPINodes"`  // 当前连接的API节点
}

type NSNodeOperator struct {
	Id                 any // ID
	AdminId            any // 管理员ID
	ClusterId          any // 集群ID
	Name               any // 节点名称
	IsOn               any // 是否启用
	Status             any // 运行状态
	UniqueId           any // 节点ID
	Secret             any // 密钥
	IsUp               any // 是否运行
	IsInstalled        any // 是否已安装
	InstallStatus      any // 安装状态
	InstallDir         any // 安装目录
	State              any // 状态
	IsActive           any // 是否活跃
	StatusIsNotified   any // 活跃状态已经通知
	InactiveNotifiedAt any // 离线通知时间
	ConnectedAPINodes  any // 当前连接的API节点
}

func NewNSNodeOperator() *NSNodeOperator {
	return &NSNodeOperator{}
}
