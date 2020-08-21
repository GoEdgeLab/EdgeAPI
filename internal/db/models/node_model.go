package models

// 节点
type Node struct {
	Id            uint32 `field:"id"`            // ID
	IsOn          uint8  `field:"isOn"`          // 是否启用
	UniqueId      string `field:"uniqueId"`      // 节点ID
	Secret        string `field:"secret"`        // 密钥
	Name          string `field:"name"`          // 节点名
	Code          string `field:"code"`          // 代号
	ClusterId     uint32 `field:"clusterId"`     // 集群ID
	RegionId      uint32 `field:"regionId"`      // 区域ID
	GroupId       uint32 `field:"groupId"`       // 分组ID
	CreatedAt     uint32 `field:"createdAt"`     // 创建时间
	Status        string `field:"status"`        // 最新的状态
	Version       uint32 `field:"version"`       // 当前版本号
	LatestVersion uint32 `field:"latestVersion"` // 最后版本号
	State         uint8  `field:"state"`         // 状态
}

type NodeOperator struct {
	Id            interface{} // ID
	IsOn          interface{} // 是否启用
	UniqueId      interface{} // 节点ID
	Secret        interface{} // 密钥
	Name          interface{} // 节点名
	Code          interface{} // 代号
	ClusterId     interface{} // 集群ID
	RegionId      interface{} // 区域ID
	GroupId       interface{} // 分组ID
	CreatedAt     interface{} // 创建时间
	Status        interface{} // 最新的状态
	Version       interface{} // 当前版本号
	LatestVersion interface{} // 最后版本号
	State         interface{} // 状态
}

func NewNodeOperator() *NodeOperator {
	return &NodeOperator{}
}
