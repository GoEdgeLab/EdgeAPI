package models

import "github.com/iwind/TeaGo/dbs"

const (
	NodeClusterFieldId                   dbs.FieldName = "id"                   // ID
	NodeClusterFieldAdminId              dbs.FieldName = "adminId"              // 管理员ID
	NodeClusterFieldUserId               dbs.FieldName = "userId"               // 用户ID
	NodeClusterFieldIsOn                 dbs.FieldName = "isOn"                 // 是否启用
	NodeClusterFieldName                 dbs.FieldName = "name"                 // 名称
	NodeClusterFieldUseAllAPINodes       dbs.FieldName = "useAllAPINodes"       // 是否使用所有API节点
	NodeClusterFieldApiNodes             dbs.FieldName = "apiNodes"             // 使用的API节点
	NodeClusterFieldInstallDir           dbs.FieldName = "installDir"           // 安装目录
	NodeClusterFieldOrder                dbs.FieldName = "order"                // 排序
	NodeClusterFieldCreatedAt            dbs.FieldName = "createdAt"            // 创建时间
	NodeClusterFieldGrantId              dbs.FieldName = "grantId"              // 默认认证方式
	NodeClusterFieldSshParams            dbs.FieldName = "sshParams"            // SSH默认参数
	NodeClusterFieldState                dbs.FieldName = "state"                // 状态
	NodeClusterFieldAutoRegister         dbs.FieldName = "autoRegister"         // 是否开启自动注册
	NodeClusterFieldUniqueId             dbs.FieldName = "uniqueId"             // 唯一ID
	NodeClusterFieldSecret               dbs.FieldName = "secret"               // 密钥
	NodeClusterFieldHealthCheck          dbs.FieldName = "healthCheck"          // 健康检查
	NodeClusterFieldDnsName              dbs.FieldName = "dnsName"              // DNS名称
	NodeClusterFieldDnsDomainId          dbs.FieldName = "dnsDomainId"          // 域名ID
	NodeClusterFieldDns                  dbs.FieldName = "dns"                  // DNS配置
	NodeClusterFieldToa                  dbs.FieldName = "toa"                  // TOA配置
	NodeClusterFieldCachePolicyId        dbs.FieldName = "cachePolicyId"        // 缓存策略ID
	NodeClusterFieldHttpFirewallPolicyId dbs.FieldName = "httpFirewallPolicyId" // WAF策略ID
	NodeClusterFieldAccessLog            dbs.FieldName = "accessLog"            // 访问日志设置
	NodeClusterFieldSystemServices       dbs.FieldName = "systemServices"       // 系统服务设置
	NodeClusterFieldTimeZone             dbs.FieldName = "timeZone"             // 时区
	NodeClusterFieldNodeMaxThreads       dbs.FieldName = "nodeMaxThreads"       // 节点最大线程数
	NodeClusterFieldDdosProtection       dbs.FieldName = "ddosProtection"       // DDoS防护设置
	NodeClusterFieldAutoOpenPorts        dbs.FieldName = "autoOpenPorts"        // 是否自动尝试开放端口
	NodeClusterFieldIsPinned             dbs.FieldName = "isPinned"             // 是否置顶
	NodeClusterFieldWebp                 dbs.FieldName = "webp"                 // WebP设置
	NodeClusterFieldUam                  dbs.FieldName = "uam"                  // UAM设置
	NodeClusterFieldClock                dbs.FieldName = "clock"                // 时钟配置
	NodeClusterFieldGlobalServerConfig   dbs.FieldName = "globalServerConfig"   // 全局服务配置
	NodeClusterFieldAutoRemoteStart      dbs.FieldName = "autoRemoteStart"      // 自动远程启动
	NodeClusterFieldAutoInstallNftables  dbs.FieldName = "autoInstallNftables"  // 自动安装nftables
	NodeClusterFieldIsAD                 dbs.FieldName = "isAD"                 // 是否为高防集群
	NodeClusterFieldHttpPages            dbs.FieldName = "httpPages"            // 自定义页面设置
	NodeClusterFieldCc                   dbs.FieldName = "cc"                   // CC设置
	NodeClusterFieldHttp3                dbs.FieldName = "http3"                // HTTP3设置
)

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
	SshParams            dbs.JSON `field:"sshParams"`            // SSH默认参数
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
	GlobalServerConfig   dbs.JSON `field:"globalServerConfig"`   // 全局服务配置
	AutoRemoteStart      bool     `field:"autoRemoteStart"`      // 自动远程启动
	AutoInstallNftables  bool     `field:"autoInstallNftables"`  // 自动安装nftables
	IsAD                 bool     `field:"isAD"`                 // 是否为高防集群
	HttpPages            dbs.JSON `field:"httpPages"`            // 自定义页面设置
	Cc                   dbs.JSON `field:"cc"`                   // CC设置
	Http3                dbs.JSON `field:"http3"`                // HTTP3设置
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
	SshParams            any // SSH默认参数
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
	GlobalServerConfig   any // 全局服务配置
	AutoRemoteStart      any // 自动远程启动
	AutoInstallNftables  any // 自动安装nftables
	IsAD                 any // 是否为高防集群
	HttpPages            any // 自定义页面设置
	Cc                   any // CC设置
	Http3                any // HTTP3设置
}

func NewNodeClusterOperator() *NodeClusterOperator {
	return &NodeClusterOperator{}
}
