package models

// 节点
type Node struct {
	Id                uint32 `field:"id"`                // ID
	AdminId           uint32 `field:"adminId"`           // 管理员ID
	UserId            uint32 `field:"userId"`            // 用户ID
	IsOn              uint8  `field:"isOn"`              // 是否启用
	IsUp              uint8  `field:"isUp"`              // 是否在线
	CountUp           uint32 `field:"countUp"`           // 连续在线次数
	CountDown         uint32 `field:"countDown"`         // 连续下线次数
	UniqueId          string `field:"uniqueId"`          // 节点ID
	Secret            string `field:"secret"`            // 密钥
	Name              string `field:"name"`              // 节点名
	Code              string `field:"code"`              // 代号
	ClusterId         uint32 `field:"clusterId"`         // 集群ID
	RegionId          uint32 `field:"regionId"`          // 区域ID
	GroupId           uint32 `field:"groupId"`           // 分组ID
	CreatedAt         uint64 `field:"createdAt"`         // 创建时间
	Status            string `field:"status"`            // 最新的状态
	Version           uint32 `field:"version"`           // 当前版本号
	LatestVersion     uint32 `field:"latestVersion"`     // 最后版本号
	InstallDir        string `field:"installDir"`        // 安装目录
	IsInstalled       uint8  `field:"isInstalled"`       // 是否已安装
	InstallStatus     string `field:"installStatus"`     // 安装状态
	State             uint8  `field:"state"`             // 状态
	ConnectedAPINodes string `field:"connectedAPINodes"` // 当前连接的API节点
	MaxCPU            uint32 `field:"maxCPU"`            // 可以使用的最多CPU
	DnsRoutes         string `field:"dnsRoutes"`         // DNS线路设置
}

type NodeOperator struct {
	Id                interface{} // ID
	AdminId           interface{} // 管理员ID
	UserId            interface{} // 用户ID
	IsOn              interface{} // 是否启用
	IsUp              interface{} // 是否在线
	CountUp           interface{} // 连续在线次数
	CountDown         interface{} // 连续下线次数
	UniqueId          interface{} // 节点ID
	Secret            interface{} // 密钥
	Name              interface{} // 节点名
	Code              interface{} // 代号
	ClusterId         interface{} // 集群ID
	RegionId          interface{} // 区域ID
	GroupId           interface{} // 分组ID
	CreatedAt         interface{} // 创建时间
	Status            interface{} // 最新的状态
	Version           interface{} // 当前版本号
	LatestVersion     interface{} // 最后版本号
	InstallDir        interface{} // 安装目录
	IsInstalled       interface{} // 是否已安装
	InstallStatus     interface{} // 安装状态
	State             interface{} // 状态
	ConnectedAPINodes interface{} // 当前连接的API节点
	MaxCPU            interface{} // 可以使用的最多CPU
	DnsRoutes         interface{} // DNS线路设置
}

func NewNodeOperator() *NodeOperator {
	return &NodeOperator{}
}
