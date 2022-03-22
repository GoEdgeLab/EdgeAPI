package models

import "github.com/iwind/TeaGo/dbs"

// NSNode 域名服务器节点
type NSNode struct {
	Id                uint32   `field:"id"`                // ID
	AdminId           uint32   `field:"adminId"`           // 管理员ID
	ClusterId         uint32   `field:"clusterId"`         // 集群ID
	Name              string   `field:"name"`              // 节点名称
	IsOn              bool     `field:"isOn"`              // 是否启用
	Status            dbs.JSON `field:"status"`            // 运行状态
	UniqueId          string   `field:"uniqueId"`          // 节点ID
	Secret            string   `field:"secret"`            // 密钥
	IsUp              uint8    `field:"isUp"`              // 是否运行
	IsInstalled       uint8    `field:"isInstalled"`       // 是否已安装
	InstallStatus     dbs.JSON `field:"installStatus"`     // 安装状态
	InstallDir        string   `field:"installDir"`        // 安装目录
	State             uint8    `field:"state"`             // 状态
	IsActive          uint8    `field:"isActive"`          // 是否活跃
	StatusIsNotified  uint8    `field:"statusIsNotified"`  // 活跃状态已经通知
	ConnectedAPINodes dbs.JSON `field:"connectedAPINodes"` // 当前连接的API节点
}

type NSNodeOperator struct {
	Id                interface{} // ID
	AdminId           interface{} // 管理员ID
	ClusterId         interface{} // 集群ID
	Name              interface{} // 节点名称
	IsOn              interface{} // 是否启用
	Status            interface{} // 运行状态
	UniqueId          interface{} // 节点ID
	Secret            interface{} // 密钥
	IsUp              interface{} // 是否运行
	IsInstalled       interface{} // 是否已安装
	InstallStatus     interface{} // 安装状态
	InstallDir        interface{} // 安装目录
	State             interface{} // 状态
	IsActive          interface{} // 是否活跃
	StatusIsNotified  interface{} // 活跃状态已经通知
	ConnectedAPINodes interface{} // 当前连接的API节点
}

func NewNSNodeOperator() *NSNodeOperator {
	return &NSNodeOperator{}
}
