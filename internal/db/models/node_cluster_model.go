package models

import "github.com/iwind/TeaGo/dbs"

// NodeCluster 节点集群
type NodeCluster struct {
	Id                   uint32   `field:"id"`                   // ID
	AdminId              uint32   `field:"adminId"`              // 管理员ID
	UserId               uint32   `field:"userId"`               // 用户ID
	IsOn                 bool     `field:"isOn"`                 // 是否启用
	Name                 string   `field:"name"`                 // 名称
	UseAllAPINodes       uint8    `field:"useAllAPINodes"`       // 是否使用所有API节点
	ApiNodes             dbs.JSON `field:"apiNodes"`             // 使用的API节点
	InstallDir           string   `field:"installDir"`           // 安装目录
	Order                uint32   `field:"order"`                // 排序
	CreatedAt            uint64   `field:"createdAt"`            // 创建时间
	GrantId              uint32   `field:"grantId"`              // 默认认证方式
	State                uint8    `field:"state"`                // 状态
	AutoRegister         uint8    `field:"autoRegister"`         // 是否开启自动注册
	UniqueId             string   `field:"uniqueId"`             // 唯一ID
	Secret               string   `field:"secret"`               // 密钥
	HealthCheck          dbs.JSON `field:"healthCheck"`          // 健康检查
	DnsName              string   `field:"dnsName"`              // DNS名称
	DnsDomainId          uint32   `field:"dnsDomainId"`          // 域名ID
	Dns                  dbs.JSON `field:"dns"`                  // DNS配置
	Toa                  dbs.JSON `field:"toa"`                  // TOA配置
	CachePolicyId        uint32   `field:"cachePolicyId"`        // 缓存策略ID
	HttpFirewallPolicyId uint32   `field:"httpFirewallPolicyId"` // WAF策略ID
	AccessLog            dbs.JSON `field:"accessLog"`            // 访问日志设置
	SystemServices       dbs.JSON `field:"systemServices"`       // 系统服务设置
	TimeZone             string   `field:"timeZone"`             // 时区
	NodeMaxThreads       uint32   `field:"nodeMaxThreads"`       // 节点最大线程数
	DdosProtection       dbs.JSON `field:"ddosProtection"`       // DDoS防护设置
	AutoOpenPorts        uint8    `field:"autoOpenPorts"`        // 是否自动尝试开放端口
	IsPinned             bool     `field:"isPinned"`             // 是否置顶
	Webp                 dbs.JSON `field:"webp"`                 // WebP设置
	Uam                  dbs.JSON `field:"uam"`                  // UAM设置
	Clock                dbs.JSON `field:"clock"`                // 时钟配置
}

type NodeClusterOperator struct {
	Id                   any // ID
	AdminId              any // 管理员ID
	UserId               any // 用户ID
	IsOn                 any // 是否启用
	Name                 any // 名称
	UseAllAPINodes       any // 是否使用所有API节点
	ApiNodes             any // 使用的API节点
	InstallDir           any // 安装目录
	Order                any // 排序
	CreatedAt            any // 创建时间
	GrantId              any // 默认认证方式
	State                any // 状态
	AutoRegister         any // 是否开启自动注册
	UniqueId             any // 唯一ID
	Secret               any // 密钥
	HealthCheck          any // 健康检查
	DnsName              any // DNS名称
	DnsDomainId          any // 域名ID
	Dns                  any // DNS配置
	Toa                  any // TOA配置
	CachePolicyId        any // 缓存策略ID
	HttpFirewallPolicyId any // WAF策略ID
	AccessLog            any // 访问日志设置
	SystemServices       any // 系统服务设置
	TimeZone             any // 时区
	NodeMaxThreads       any // 节点最大线程数
	DdosProtection       any // DDoS防护设置
	AutoOpenPorts        any // 是否自动尝试开放端口
	IsPinned             any // 是否置顶
	Webp                 any // WebP设置
	Uam                  any // UAM设置
	Clock                any // 时钟配置
}

func NewNodeClusterOperator() *NodeClusterOperator {
	return &NodeClusterOperator{}
}
