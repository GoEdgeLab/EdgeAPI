package models

import "github.com/iwind/TeaGo/dbs"

// NodeIPAddress 节点IP地址
type NodeIPAddress struct {
	Id                uint32   `field:"id"`                // ID
	NodeId            uint32   `field:"nodeId"`            // 节点ID
	Role              string   `field:"role"`              // 节点角色
	GroupId           uint32   `field:"groupId"`           // 所属分组ID
	Name              string   `field:"name"`              // 名称
	Ip                string   `field:"ip"`                // IP地址
	Description       string   `field:"description"`       // 描述
	State             uint8    `field:"state"`             // 状态
	Order             uint32   `field:"order"`             // 排序
	CanAccess         uint8    `field:"canAccess"`         // 是否可以访问
	IsOn              bool     `field:"isOn"`              // 是否启用
	IsUp              uint8    `field:"isUp"`              // 是否上线
	IsHealthy         uint8    `field:"isHealthy"`         // 是否健康
	Thresholds        dbs.JSON `field:"thresholds"`        // 上线阈值
	Connectivity      dbs.JSON `field:"connectivity"`      // 连通性状态
	BackupIP          string   `field:"backupIP"`          // 备用IP
	BackupThresholdId uint32   `field:"backupThresholdId"` // 触发备用IP的阈值
	CountUp           uint32   `field:"countUp"`           // UP状态次数
	CountDown         uint32   `field:"countDown"`         // DOWN状态次数
}

type NodeIPAddressOperator struct {
	Id                interface{} // ID
	NodeId            interface{} // 节点ID
	Role              interface{} // 节点角色
	GroupId           interface{} // 所属分组ID
	Name              interface{} // 名称
	Ip                interface{} // IP地址
	Description       interface{} // 描述
	State             interface{} // 状态
	Order             interface{} // 排序
	CanAccess         interface{} // 是否可以访问
	IsOn              interface{} // 是否启用
	IsUp              interface{} // 是否上线
	IsHealthy         interface{} // 是否健康
	Thresholds        interface{} // 上线阈值
	Connectivity      interface{} // 连通性状态
	BackupIP          interface{} // 备用IP
	BackupThresholdId interface{} // 触发备用IP的阈值
	CountUp           interface{} // UP状态次数
	CountDown         interface{} // DOWN状态次数
}

func NewNodeIPAddressOperator() *NodeIPAddressOperator {
	return &NodeIPAddressOperator{}
}
