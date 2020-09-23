package models

// 节点集群
type NodeCluster struct {
	Id             uint32 `field:"id"`             // ID
	AdminId        uint32 `field:"adminId"`        // 管理员ID
	UserId         uint32 `field:"userId"`         // 用户ID
	Name           string `field:"name"`           // 名称
	UseAllAPINodes uint8  `field:"useAllAPINodes"` // 是否使用所有API节点
	ApiNodes       string `field:"apiNodes"`       // 使用的API节点
	InstallDir     string `field:"installDir"`     // 安装目录
	Order          uint32 `field:"order"`          // 排序
	CreatedAt      uint64 `field:"createdAt"`      // 创建时间
	GrantId        uint32 `field:"grantId"`        // 默认认证方式
	State          uint8  `field:"state"`          // 状态
}

type NodeClusterOperator struct {
	Id             interface{} // ID
	AdminId        interface{} // 管理员ID
	UserId         interface{} // 用户ID
	Name           interface{} // 名称
	UseAllAPINodes interface{} // 是否使用所有API节点
	ApiNodes       interface{} // 使用的API节点
	InstallDir     interface{} // 安装目录
	Order          interface{} // 排序
	CreatedAt      interface{} // 创建时间
	GrantId        interface{} // 默认认证方式
	State          interface{} // 状态
}

func NewNodeClusterOperator() *NodeClusterOperator {
	return &NodeClusterOperator{}
}
