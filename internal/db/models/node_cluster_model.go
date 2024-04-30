package models

import "github.com/iwind/TeaGo/dbs"

const (
	NodeClusterField_Id                   dbs.FieldName = "id"                   // ID
	NodeClusterField_AdminId              dbs.FieldName = "adminId"              // 管理员ID
	NodeClusterField_UserId               dbs.FieldName = "userId"               // 用户ID
	NodeClusterField_IsOn                 dbs.FieldName = "isOn"                 // 是否启用
	NodeClusterField_Name                 dbs.FieldName = "name"                 // 名称
	NodeClusterField_UseAllAPINodes       dbs.FieldName = "useAllAPINodes"       // 是否使用所有API节点
	NodeClusterField_ApiNodes             dbs.FieldName = "apiNodes"             // 使用的API节点
	NodeClusterField_InstallDir           dbs.FieldName = "installDir"           // 安装目录
	NodeClusterField_Order                dbs.FieldName = "order"                // 排序
	NodeClusterField_CreatedAt            dbs.FieldName = "createdAt"            // 创建时间
	NodeClusterField_GrantId              dbs.FieldName = "grantId"              // 默认认证方式
	NodeClusterField_SshParams            dbs.FieldName = "sshParams"            // SSH默认参数
	NodeClusterField_State                dbs.FieldName = "state"                // 状态
	NodeClusterField_AutoRegister         dbs.FieldName = "autoRegister"         // 是否开启自动注册
	NodeClusterField_UniqueId             dbs.FieldName = "uniqueId"             // 唯一ID
	NodeClusterField_Secret               dbs.FieldName = "secret"               // 密钥
	NodeClusterField_HealthCheck          dbs.FieldName = "healthCheck"          // 健康检查
	NodeClusterField_DnsName              dbs.FieldName = "dnsName"              // DNS名称
	NodeClusterField_DnsDomainId          dbs.FieldName = "dnsDomainId"          // 域名ID
	NodeClusterField_Dns                  dbs.FieldName = "dns"                  // DNS配置
	NodeClusterField_Toa                  dbs.FieldName = "toa"                  // TOA配置
	NodeClusterField_CachePolicyId        dbs.FieldName = "cachePolicyId"        // 缓存策略ID
	NodeClusterField_HttpFirewallPolicyId dbs.FieldName = "httpFirewallPolicyId" // WAF策略ID
	NodeClusterField_AccessLog            dbs.FieldName = "accessLog"            // 访问日志设置
	NodeClusterField_SystemServices       dbs.FieldName = "systemServices"       // 系统服务设置
	NodeClusterField_TimeZone             dbs.FieldName = "timeZone"             // 时区
	NodeClusterField_NodeMaxThreads       dbs.FieldName = "nodeMaxThreads"       // 节点最大线程数
	NodeClusterField_DdosProtection       dbs.FieldName = "ddosProtection"       // DDoS防护设置
	NodeClusterField_AutoOpenPorts        dbs.FieldName = "autoOpenPorts"        // 是否自动尝试开放端口
	NodeClusterField_IsPinned             dbs.FieldName = "isPinned"             // 是否置顶
	NodeClusterField_Webp                 dbs.FieldName = "webp"                 // WebP设置
	NodeClusterField_Uam                  dbs.FieldName = "uam"                  // UAM设置
	NodeClusterField_Clock                dbs.FieldName = "clock"                // 时钟配置
	NodeClusterField_GlobalServerConfig   dbs.FieldName = "globalServerConfig"   // 全局服务配置
	NodeClusterField_AutoRemoteStart      dbs.FieldName = "autoRemoteStart"      // 自动远程启动
	NodeClusterField_AutoInstallNftables  dbs.FieldName = "autoInstallNftables"  // 自动安装nftables
	NodeClusterField_IsAD                 dbs.FieldName = "isAD"                 // 是否为高防集群
	NodeClusterField_HttpPages            dbs.FieldName = "httpPages"            // 自定义页面设置
	NodeClusterField_Cc                   dbs.FieldName = "cc"                   // CC设置
	NodeClusterField_Http3                dbs.FieldName = "http3"                // HTTP3设置
	NodeClusterField_AutoSystemTuning     dbs.FieldName = "autoSystemTuning"     // 是否自动调整系统参数
	NodeClusterField_NetworkSecurity      dbs.FieldName = "networkSecurity"      // 网络安全策略
	NodeClusterField_AutoTrimDisks        dbs.FieldName = "autoTrimDisks"        // 是否自动执行TRIM
	NodeClusterField_MaxConcurrentReads   dbs.FieldName = "maxConcurrentReads"   // 节点并发读限制
	NodeClusterField_MaxConcurrentWrites  dbs.FieldName = "maxConcurrentWrites"  // 节点并发写限制
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
	AutoSystemTuning     bool     `field:"autoSystemTuning"`     // 是否自动调整系统参数
	NetworkSecurity      dbs.JSON `field:"networkSecurity"`      // 网络安全策略
	AutoTrimDisks        bool     `field:"autoTrimDisks"`        // 是否自动执行TRIM
	MaxConcurrentReads   uint32   `field:"maxConcurrentReads"`   // 节点并发读限制
	MaxConcurrentWrites  uint32   `field:"maxConcurrentWrites"`  // 节点并发写限制
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
	AutoSystemTuning     any // 是否自动调整系统参数
	NetworkSecurity      any // 网络安全策略
	AutoTrimDisks        any // 是否自动执行TRIM
	MaxConcurrentReads   any // 节点并发读限制
	MaxConcurrentWrites  any // 节点并发写限制
}

func NewNodeClusterOperator() *NodeClusterOperator {
	return &NodeClusterOperator{}
}
