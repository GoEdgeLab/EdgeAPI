package models

import "github.com/iwind/TeaGo/dbs"

// Node 节点
type Node struct {
	Id                     uint32   `field:"id"`                     // ID
	AdminId                uint32   `field:"adminId"`                // 管理员ID
	UserId                 uint32   `field:"userId"`                 // 用户ID
	Level                  uint8    `field:"level"`                  // 级别
	LnAddrs                dbs.JSON `field:"lnAddrs"`                // Ln级别访问地址
	IsOn                   bool     `field:"isOn"`                   // 是否启用
	IsUp                   bool     `field:"isUp"`                   // 是否在线
	CountUp                uint32   `field:"countUp"`                // 连续在线次数
	CountDown              uint32   `field:"countDown"`              // 连续下线次数
	IsActive               bool     `field:"isActive"`               // 是否活跃
	InactiveNotifiedAt     uint64   `field:"inactiveNotifiedAt"`     // 离线通知时间
	UniqueId               string   `field:"uniqueId"`               // 节点ID
	Secret                 string   `field:"secret"`                 // 密钥
	Name                   string   `field:"name"`                   // 节点名
	Code                   string   `field:"code"`                   // 代号
	ClusterId              uint32   `field:"clusterId"`              // 主集群ID
	SecondaryClusterIds    dbs.JSON `field:"secondaryClusterIds"`    // 从集群ID
	RegionId               uint32   `field:"regionId"`               // 区域ID
	GroupId                uint32   `field:"groupId"`                // 分组ID
	CreatedAt              uint64   `field:"createdAt"`              // 创建时间
	Status                 dbs.JSON `field:"status"`                 // 最新的状态
	Version                uint32   `field:"version"`                // 当前版本号
	LatestVersion          uint32   `field:"latestVersion"`          // 最后版本号
	InstallDir             string   `field:"installDir"`             // 安装目录
	IsInstalled            bool     `field:"isInstalled"`            // 是否已安装
	InstallStatus          dbs.JSON `field:"installStatus"`          // 安装状态
	State                  uint8    `field:"state"`                  // 状态
	ConnectedAPINodes      dbs.JSON `field:"connectedAPINodes"`      // 当前连接的API节点
	MaxCPU                 uint32   `field:"maxCPU"`                 // 可以使用的最多CPU
	MaxThreads             uint32   `field:"maxThreads"`             // 最大线程数
	DdosProtection         dbs.JSON `field:"ddosProtection"`         // DDOS配置
	DnsRoutes              dbs.JSON `field:"dnsRoutes"`              // DNS线路设置
	MaxCacheDiskCapacity   dbs.JSON `field:"maxCacheDiskCapacity"`   // 硬盘缓存容量
	MaxCacheMemoryCapacity dbs.JSON `field:"maxCacheMemoryCapacity"` // 内存缓存容量
	CacheDiskDir           string   `field:"cacheDiskDir"`           // 缓存目录
	DnsResolver            dbs.JSON `field:"dnsResolver"`            // DNS解析器
	EnableIPLists          bool     `field:"enableIPLists"`          // 启用IP名单
}

type NodeOperator struct {
	Id                     any // ID
	AdminId                any // 管理员ID
	UserId                 any // 用户ID
	Level                  any // 级别
	LnAddrs                any // Ln级别访问地址
	IsOn                   any // 是否启用
	IsUp                   any // 是否在线
	CountUp                any // 连续在线次数
	CountDown              any // 连续下线次数
	IsActive               any // 是否活跃
	InactiveNotifiedAt     any // 离线通知时间
	UniqueId               any // 节点ID
	Secret                 any // 密钥
	Name                   any // 节点名
	Code                   any // 代号
	ClusterId              any // 主集群ID
	SecondaryClusterIds    any // 从集群ID
	RegionId               any // 区域ID
	GroupId                any // 分组ID
	CreatedAt              any // 创建时间
	Status                 any // 最新的状态
	Version                any // 当前版本号
	LatestVersion          any // 最后版本号
	InstallDir             any // 安装目录
	IsInstalled            any // 是否已安装
	InstallStatus          any // 安装状态
	State                  any // 状态
	ConnectedAPINodes      any // 当前连接的API节点
	MaxCPU                 any // 可以使用的最多CPU
	MaxThreads             any // 最大线程数
	DdosProtection         any // DDOS配置
	DnsRoutes              any // DNS线路设置
	MaxCacheDiskCapacity   any // 硬盘缓存容量
	MaxCacheMemoryCapacity any // 内存缓存容量
	CacheDiskDir           any // 缓存目录
	DnsResolver            any // DNS解析器
	EnableIPLists          any // 启用IP名单
}

func NewNodeOperator() *NodeOperator {
	return &NodeOperator{}
}
