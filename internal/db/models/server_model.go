package models

import "github.com/iwind/TeaGo/dbs"

const (
	ServerField_Id                  dbs.FieldName = "id"                  // ID
	ServerField_IsOn                dbs.FieldName = "isOn"                // 是否启用
	ServerField_UserId              dbs.FieldName = "userId"              // 用户ID
	ServerField_AdminId             dbs.FieldName = "adminId"             // 管理员ID
	ServerField_Type                dbs.FieldName = "type"                // 服务类型
	ServerField_Name                dbs.FieldName = "name"                // 名称
	ServerField_Description         dbs.FieldName = "description"         // 描述
	ServerField_PlainServerNames    dbs.FieldName = "plainServerNames"    // 扁平化域名列表
	ServerField_ServerNames         dbs.FieldName = "serverNames"         // 域名列表
	ServerField_AuditingAt          dbs.FieldName = "auditingAt"          // 审核提交时间
	ServerField_AuditingServerNames dbs.FieldName = "auditingServerNames" // 审核中的域名
	ServerField_IsAuditing          dbs.FieldName = "isAuditing"          // 是否正在审核
	ServerField_AuditingResult      dbs.FieldName = "auditingResult"      // 审核结果
	ServerField_Http                dbs.FieldName = "http"                // HTTP配置
	ServerField_Https               dbs.FieldName = "https"               // HTTPS配置
	ServerField_Tcp                 dbs.FieldName = "tcp"                 // TCP配置
	ServerField_Tls                 dbs.FieldName = "tls"                 // TLS配置
	ServerField_Unix                dbs.FieldName = "unix"                // Unix配置
	ServerField_Udp                 dbs.FieldName = "udp"                 // UDP配置
	ServerField_WebId               dbs.FieldName = "webId"               // WEB配置
	ServerField_ReverseProxy        dbs.FieldName = "reverseProxy"        // 反向代理配置
	ServerField_GroupIds            dbs.FieldName = "groupIds"            // 分组ID列表
	ServerField_Config              dbs.FieldName = "config"              // 服务配置，自动生成
	ServerField_ConfigMd5           dbs.FieldName = "configMd5"           // Md5
	ServerField_ClusterId           dbs.FieldName = "clusterId"           // 集群ID
	ServerField_IncludeNodes        dbs.FieldName = "includeNodes"        // 部署条件
	ServerField_ExcludeNodes        dbs.FieldName = "excludeNodes"        // 节点排除条件
	ServerField_Version             dbs.FieldName = "version"             // 版本号
	ServerField_CreatedAt           dbs.FieldName = "createdAt"           // 创建时间
	ServerField_State               dbs.FieldName = "state"               // 状态
	ServerField_DnsName             dbs.FieldName = "dnsName"             // DNS名称
	ServerField_TcpPorts            dbs.FieldName = "tcpPorts"            // 所包含TCP端口
	ServerField_UdpPorts            dbs.FieldName = "udpPorts"            // 所包含UDP端口
	ServerField_SupportCNAME        dbs.FieldName = "supportCNAME"        // 允许CNAME不在域名名单
	ServerField_TrafficLimit        dbs.FieldName = "trafficLimit"        // 流量限制
	ServerField_TrafficDay          dbs.FieldName = "trafficDay"          // YYYYMMDD
	ServerField_TrafficMonth        dbs.FieldName = "trafficMonth"        // YYYYMM
	ServerField_TotalDailyTraffic   dbs.FieldName = "totalDailyTraffic"   // 日流量
	ServerField_TotalMonthlyTraffic dbs.FieldName = "totalMonthlyTraffic" // 月流量
	ServerField_TrafficLimitStatus  dbs.FieldName = "trafficLimitStatus"  // 流量限制状态
	ServerField_TotalTraffic        dbs.FieldName = "totalTraffic"        // 总流量
	ServerField_UserPlanId          dbs.FieldName = "userPlanId"          // 所属套餐ID
	ServerField_LastUserPlanId      dbs.FieldName = "lastUserPlanId"      // 上一次使用的套餐
	ServerField_Uam                 dbs.FieldName = "uam"                 // UAM设置
	ServerField_BandwidthTime       dbs.FieldName = "bandwidthTime"       // 带宽更新时间，YYYYMMDDHHII
	ServerField_BandwidthBytes      dbs.FieldName = "bandwidthBytes"      // 最近带宽峰值
	ServerField_CountAttackRequests dbs.FieldName = "countAttackRequests" // 最近攻击请求数
	ServerField_CountRequests       dbs.FieldName = "countRequests"       // 最近总请求数
)

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
	CountAttackRequests uint64   `field:"countAttackRequests"` // 最近攻击请求数
	CountRequests       uint64   `field:"countRequests"`       // 最近总请求数
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
	CountAttackRequests any // 最近攻击请求数
	CountRequests       any // 最近总请求数
}

func NewServerOperator() *ServerOperator {
	return &ServerOperator{}
}
