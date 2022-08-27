package models

import "github.com/iwind/TeaGo/dbs"

// Server 服务
type Server struct {
	Id                  uint32   `field:"id"`                  // ID
	IsOn                bool     `field:"isOn"`                // 是否启用
	UserId              uint32   `field:"userId"`              // 用户ID
	AdminId             uint32   `field:"adminId"`             // 管理员ID
	Type                string   `field:"type"`                // 服务类型
	Name                string   `field:"name"`                // 名称
	Description         string   `field:"description"`         // 描述
	PlainServerNames    dbs.JSON `field:"plainServerNames"`    // 扁平化域名列表
	ServerNames         dbs.JSON `field:"serverNames"`         // 域名列表
	AuditingAt          uint64   `field:"auditingAt"`          // 审核提交时间
	AuditingServerNames dbs.JSON `field:"auditingServerNames"` // 审核中的域名
	IsAuditing          bool     `field:"isAuditing"`          // 是否正在审核
	AuditingResult      dbs.JSON `field:"auditingResult"`      // 审核结果
	Http                dbs.JSON `field:"http"`                // HTTP配置
	Https               dbs.JSON `field:"https"`               // HTTPS配置
	Tcp                 dbs.JSON `field:"tcp"`                 // TCP配置
	Tls                 dbs.JSON `field:"tls"`                 // TLS配置
	Unix                dbs.JSON `field:"unix"`                // Unix配置
	Udp                 dbs.JSON `field:"udp"`                 // UDP配置
	WebId               uint32   `field:"webId"`               // WEB配置
	ReverseProxy        dbs.JSON `field:"reverseProxy"`        // 反向代理配置
	GroupIds            dbs.JSON `field:"groupIds"`            // 分组ID列表
	Config              dbs.JSON `field:"config"`              // 服务配置，自动生成
	ConfigMd5           string   `field:"configMd5"`           // Md5
	ClusterId           uint32   `field:"clusterId"`           // 集群ID
	IncludeNodes        dbs.JSON `field:"includeNodes"`        // 部署条件
	ExcludeNodes        dbs.JSON `field:"excludeNodes"`        // 节点排除条件
	Version             uint32   `field:"version"`             // 版本号
	CreatedAt           uint64   `field:"createdAt"`           // 创建时间
	State               uint8    `field:"state"`               // 状态
	DnsName             string   `field:"dnsName"`             // DNS名称
	TcpPorts            dbs.JSON `field:"tcpPorts"`            // 所包含TCP端口
	UdpPorts            dbs.JSON `field:"udpPorts"`            // 所包含UDP端口
	SupportCNAME        uint8    `field:"supportCNAME"`        // 允许CNAME不在域名名单
	TrafficLimit        dbs.JSON `field:"trafficLimit"`        // 流量限制
	TrafficDay          string   `field:"trafficDay"`          // YYYYMMDD
	TrafficMonth        string   `field:"trafficMonth"`        // YYYYMM
	TotalDailyTraffic   float64  `field:"totalDailyTraffic"`   // 日流量
	TotalMonthlyTraffic float64  `field:"totalMonthlyTraffic"` // 月流量
	TrafficLimitStatus  dbs.JSON `field:"trafficLimitStatus"`  // 流量限制状态
	TotalTraffic        float64  `field:"totalTraffic"`        // 总流量
	UserPlanId          uint32   `field:"userPlanId"`          // 所属套餐ID
	LastUserPlanId      uint32   `field:"lastUserPlanId"`      // 上一次使用的套餐
	Uam                 dbs.JSON `field:"uam"`                 // UAM设置
	BandwidthTime       string   `field:"bandwidthTime"`       // 带宽更新时间，YYYYMMDDHHII
	BandwidthBytes      uint64   `field:"bandwidthBytes"`      // 最近带宽峰值
}

type ServerOperator struct {
	Id                  any // ID
	IsOn                any // 是否启用
	UserId              any // 用户ID
	AdminId             any // 管理员ID
	Type                any // 服务类型
	Name                any // 名称
	Description         any // 描述
	PlainServerNames    any // 扁平化域名列表
	ServerNames         any // 域名列表
	AuditingAt          any // 审核提交时间
	AuditingServerNames any // 审核中的域名
	IsAuditing          any // 是否正在审核
	AuditingResult      any // 审核结果
	Http                any // HTTP配置
	Https               any // HTTPS配置
	Tcp                 any // TCP配置
	Tls                 any // TLS配置
	Unix                any // Unix配置
	Udp                 any // UDP配置
	WebId               any // WEB配置
	ReverseProxy        any // 反向代理配置
	GroupIds            any // 分组ID列表
	Config              any // 服务配置，自动生成
	ConfigMd5           any // Md5
	ClusterId           any // 集群ID
	IncludeNodes        any // 部署条件
	ExcludeNodes        any // 节点排除条件
	Version             any // 版本号
	CreatedAt           any // 创建时间
	State               any // 状态
	DnsName             any // DNS名称
	TcpPorts            any // 所包含TCP端口
	UdpPorts            any // 所包含UDP端口
	SupportCNAME        any // 允许CNAME不在域名名单
	TrafficLimit        any // 流量限制
	TrafficDay          any // YYYYMMDD
	TrafficMonth        any // YYYYMM
	TotalDailyTraffic   any // 日流量
	TotalMonthlyTraffic any // 月流量
	TrafficLimitStatus  any // 流量限制状态
	TotalTraffic        any // 总流量
	UserPlanId          any // 所属套餐ID
	LastUserPlanId      any // 上一次使用的套餐
	Uam                 any // UAM设置
	BandwidthTime       any // 带宽更新时间，YYYYMMDDHHII
	BandwidthBytes      any // 最近带宽峰值
}

func NewServerOperator() *ServerOperator {
	return &ServerOperator{}
}
