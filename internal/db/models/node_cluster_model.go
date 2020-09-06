package models

// 节点集群
type NodeCluster struct {
	Id             uint32 `field:"id"`             // ID
	Name           string `field:"name"`           // 名称
	InstallDir     string `field:"installDir"`     // 安装目录
	GrantId        uint32 `field:"grantId"`        // 默认认证方式
	UseAllAPINodes uint8  `field:"useAllAPINodes"` // 是否使用所有API节点
	ApiNodes       string `field:"apiNodes"`       // 使用的API节点
	Order          uint32 `field:"order"`          // 排序
	CreatedAt      uint32 `field:"createdAt"`      // 创建时间
	State          uint8  `field:"state"`          // 状态
}

type NodeClusterOperator struct {
	Id             interface{} // ID
	Name           interface{} // 名称
	InstallDir     interface{} // 安装目录
	GrantId        interface{} // 默认认证方式
	UseAllAPINodes interface{} // 是否使用所有API节点
	ApiNodes       interface{} // 使用的API节点
	Order          interface{} // 排序
	CreatedAt      interface{} // 创建时间
	State          interface{} // 状态
}

func NewNodeClusterOperator() *NodeClusterOperator {
	return &NodeClusterOperator{}
}
