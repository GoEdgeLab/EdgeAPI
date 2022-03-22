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
	ServerNames         dbs.JSON `field:"serverNames"`         // 域名列表
	AuditingAt          uint64   `field:"auditingAt"`          // 审核提交时间
	AuditingServerNames dbs.JSON `field:"auditingServerNames"` // 审核中的域名
	IsAuditing          uint8    `field:"isAuditing"`          // 是否正在审核
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
}

type ServerOperator struct {
	Id                  interface{} // ID
	IsOn                interface{} // 是否启用
	UserId              interface{} // 用户ID
	AdminId             interface{} // 管理员ID
	Type                interface{} // 服务类型
	Name                interface{} // 名称
	Description         interface{} // 描述
	ServerNames         interface{} // 域名列表
	AuditingAt          interface{} // 审核提交时间
	AuditingServerNames interface{} // 审核中的域名
	IsAuditing          interface{} // 是否正在审核
	AuditingResult      interface{} // 审核结果
	Http                interface{} // HTTP配置
	Https               interface{} // HTTPS配置
	Tcp                 interface{} // TCP配置
	Tls                 interface{} // TLS配置
	Unix                interface{} // Unix配置
	Udp                 interface{} // UDP配置
	WebId               interface{} // WEB配置
	ReverseProxy        interface{} // 反向代理配置
	GroupIds            interface{} // 分组ID列表
	Config              interface{} // 服务配置，自动生成
	ConfigMd5           interface{} // Md5
	ClusterId           interface{} // 集群ID
	IncludeNodes        interface{} // 部署条件
	ExcludeNodes        interface{} // 节点排除条件
	Version             interface{} // 版本号
	CreatedAt           interface{} // 创建时间
	State               interface{} // 状态
	DnsName             interface{} // DNS名称
	TcpPorts            interface{} // 所包含TCP端口
	UdpPorts            interface{} // 所包含UDP端口
	SupportCNAME        interface{} // 允许CNAME不在域名名单
	TrafficLimit        interface{} // 流量限制
	TrafficDay          interface{} // YYYYMMDD
	TrafficMonth        interface{} // YYYYMM
	TotalDailyTraffic   interface{} // 日流量
	TotalMonthlyTraffic interface{} // 月流量
	TrafficLimitStatus  interface{} // 流量限制状态
	TotalTraffic        interface{} // 总流量
	UserPlanId          interface{} // 所属套餐ID
	LastUserPlanId      interface{} // 上一次使用的套餐
}

func NewServerOperator() *ServerOperator {
	return &ServerOperator{}
}
