package models

// 服务
type Server struct {
	Id           uint32 `field:"id"`           // ID
	UniqueId     string `field:"uniqueId"`     // 唯一ID
	IsOn         uint8  `field:"isOn"`         // 是否启用
	UserId       uint32 `field:"userId"`       // 用户ID
	AdminId      uint32 `field:"adminId"`      // 管理员ID
	GroupIds     string `field:"groupIds"`     // 分组ID列表
	Config       string `field:"config"`       // 服务配置，自动生成
	ClusterId    uint32 `field:"clusterId"`    // 集群ID
	IncludeNodes string `field:"includeNodes"` // 部署条件
	ExcludeNodes string `field:"excludeNodes"` // 节点排除条件
	Version      uint32 `field:"version"`      // 版本号
	CreatedAt    uint32 `field:"createdAt"`    // 创建时间
	State        uint8  `field:"state"`        // 状态
}

type ServerOperator struct {
	Id           interface{} // ID
	UniqueId     interface{} // 唯一ID
	IsOn         interface{} // 是否启用
	UserId       interface{} // 用户ID
	AdminId      interface{} // 管理员ID
	GroupIds     interface{} // 分组ID列表
	Config       interface{} // 服务配置，自动生成
	ClusterId    interface{} // 集群ID
	IncludeNodes interface{} // 部署条件
	ExcludeNodes interface{} // 节点排除条件
	Version      interface{} // 版本号
	CreatedAt    interface{} // 创建时间
	State        interface{} // 状态
}

func NewServerOperator() *ServerOperator {
	return &ServerOperator{}
}
