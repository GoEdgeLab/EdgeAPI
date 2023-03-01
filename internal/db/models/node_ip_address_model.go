package models

import "github.com/iwind/TeaGo/dbs"

// NodeIPAddress 节点IP地址
type NodeIPAddress struct {
	Id                uint32   `field:"id"`                // ID
	NodeId            uint32   `field:"nodeId"`            // 节点ID
	ClusterIds        dbs.JSON `field:"clusterIds"`        // 所属集群IDs
	Role              string   `field:"role"`              // 节点角色
	GroupId           uint32   `field:"groupId"`           // 所属分组ID
	Name              string   `field:"name"`              // 名称
	Ip                string   `field:"ip"`                // IP地址
	Description       string   `field:"description"`       // 描述
	State             uint8    `field:"state"`             // 状态
	Order             uint32   `field:"order"`             // 排序
	CanAccess         bool     `field:"canAccess"`         // 是否可以访问
	IsOn              bool     `field:"isOn"`              // 是否启用
	IsUp              bool     `field:"isUp"`              // 是否上线
	IsHealthy         bool     `field:"isHealthy"`         // 是否健康
	Thresholds        dbs.JSON `field:"thresholds"`        // 上线阈值
	Connectivity      dbs.JSON `field:"connectivity"`      // 连通性状态
	BackupIP          string   `field:"backupIP"`          // 备用IP
	BackupThresholdId uint32   `field:"backupThresholdId"` // 触发备用IP的阈值
	CountUp           uint32   `field:"countUp"`           // UP状态次数
	CountDown         uint32   `field:"countDown"`         // DOWN状态次数
}

type NodeIPAddressOperator struct {
	Id                any // ID
	NodeId            any // 节点ID
	ClusterIds        any // 所属集群IDs
	Role              any // 节点角色
	GroupId           any // 所属分组ID
	Name              any // 名称
	Ip                any // IP地址
	Description       any // 描述
	State             any // 状态
	Order             any // 排序
	CanAccess         any // 是否可以访问
	IsOn              any // 是否启用
	IsUp              any // 是否上线
	IsHealthy         any // 是否健康
	Thresholds        any // 上线阈值
	Connectivity      any // 连通性状态
	BackupIP          any // 备用IP
	BackupThresholdId any // 触发备用IP的阈值
	CountUp           any // UP状态次数
	CountDown         any // DOWN状态次数
}

func NewNodeIPAddressOperator() *NodeIPAddressOperator {
	return &NodeIPAddressOperator{}
}
