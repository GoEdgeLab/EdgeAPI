package models

// 节点集群
type NodeCluster struct {
	Id                   uint32 `field:"id"`                   // ID
	AdminId              uint32 `field:"adminId"`              // 管理员ID
	UserId               uint32 `field:"userId"`               // 用户ID
	Name                 string `field:"name"`                 // 名称
	UseAllAPINodes       uint8  `field:"useAllAPINodes"`       // 是否使用所有API节点
	ApiNodes             string `field:"apiNodes"`             // 使用的API节点
	InstallDir           string `field:"installDir"`           // 安装目录
	Order                uint32 `field:"order"`                // 排序
	CreatedAt            uint64 `field:"createdAt"`            // 创建时间
	GrantId              uint32 `field:"grantId"`              // 默认认证方式
	State                uint8  `field:"state"`                // 状态
	AutoRegister         uint8  `field:"autoRegister"`         // 是否开启自动注册
	UniqueId             string `field:"uniqueId"`             // 唯一ID
	Secret               string `field:"secret"`               // 密钥
	HealthCheck          string `field:"healthCheck"`          // 健康检查
	DnsName              string `field:"dnsName"`              // DNS名称
	DnsDomainId          uint32 `field:"dnsDomainId"`          // 域名ID
	Dns                  string `field:"dns"`                  // DNS配置
	Toa                  string `field:"toa"`                  // TOA配置
	CachePolicyId        uint32 `field:"cachePolicyId"`        // 缓存策略ID
	HttpFirewallPolicyId uint32 `field:"httpFirewallPolicyId"` // WAF策略ID
	AccessLog            string `field:"accessLog"`            // 访问日志设置
	SystemServices       string `field:"systemServices"`       // 系统服务设置
}

type NodeClusterOperator struct {
	Id                   interface{} // ID
	AdminId              interface{} // 管理员ID
	UserId               interface{} // 用户ID
	Name                 interface{} // 名称
	UseAllAPINodes       interface{} // 是否使用所有API节点
	ApiNodes             interface{} // 使用的API节点
	InstallDir           interface{} // 安装目录
	Order                interface{} // 排序
	CreatedAt            interface{} // 创建时间
	GrantId              interface{} // 默认认证方式
	State                interface{} // 状态
	AutoRegister         interface{} // 是否开启自动注册
	UniqueId             interface{} // 唯一ID
	Secret               interface{} // 密钥
	HealthCheck          interface{} // 健康检查
	DnsName              interface{} // DNS名称
	DnsDomainId          interface{} // 域名ID
	Dns                  interface{} // DNS配置
	Toa                  interface{} // TOA配置
	CachePolicyId        interface{} // 缓存策略ID
	HttpFirewallPolicyId interface{} // WAF策略ID
	AccessLog            interface{} // 访问日志设置
	SystemServices       interface{} // 系统服务设置
}

func NewNodeClusterOperator() *NodeClusterOperator {
	return &NodeClusterOperator{}
}
