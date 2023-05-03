package nameservers

import "github.com/iwind/TeaGo/dbs"

// NSRecord DNS记录
type NSRecord struct {
	Id          uint64   `field:"id"`          // ID
	DomainId    uint32   `field:"domainId"`    // 域名ID
	IsOn        bool     `field:"isOn"`        // 是否启用
	Description string   `field:"description"` // 备注
	Name        string   `field:"name"`        // 记录名
	Type        string   `field:"type"`        // 类型
	Value       string   `field:"value"`       // 值
	MxPriority  uint32   `field:"mxPriority"`  // MX优先级
	SrvPriority uint32   `field:"srvPriority"` // SRV优先级
	SrvWeight   uint32   `field:"srvWeight"`   // SRV权重
	SrvPort     uint32   `field:"srvPort"`     // SRV端口
	CaaFlag     uint8    `field:"caaFlag"`     // CAA Flag
	CaaTag      string   `field:"caaTag"`      // CAA TAG
	Ttl         uint32   `field:"ttl"`         // TTL（秒）
	Weight      uint32   `field:"weight"`      // 权重
	RouteIds    dbs.JSON `field:"routeIds"`    // 线路
	HealthCheck dbs.JSON `field:"healthCheck"` // 健康检查配置
	CountUp     uint32   `field:"countUp"`     // 连续上线次数
	CountDown   uint32   `field:"countDown"`   // 连续离线次数
	IsUp        bool     `field:"isUp"`        // 是否在线
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
	Version     uint64   `field:"version"`     // 版本号
	State       uint8    `field:"state"`       // 状态
}

type NSRecordOperator struct {
	Id          any // ID
	DomainId    any // 域名ID
	IsOn        any // 是否启用
	Description any // 备注
	Name        any // 记录名
	Type        any // 类型
	Value       any // 值
	MxPriority  any // MX优先级
	SrvPriority any // SRV优先级
	SrvWeight   any // SRV权重
	SrvPort     any // SRV端口
	CaaFlag     any // CAA Flag
	CaaTag      any // CAA TAG
	Ttl         any // TTL（秒）
	Weight      any // 权重
	RouteIds    any // 线路
	HealthCheck any // 健康检查配置
	CountUp     any // 连续上线次数
	CountDown   any // 连续离线次数
	IsUp        any // 是否在线
	CreatedAt   any // 创建时间
	Version     any // 版本号
	State       any // 状态
}

func NewNSRecordOperator() *NSRecordOperator {
	return &NSRecordOperator{}
}
