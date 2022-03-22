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
	Ttl         uint32   `field:"ttl"`         // TTL（秒）
	Weight      uint32   `field:"weight"`      // 权重
	RouteIds    dbs.JSON `field:"routeIds"`    // 线路
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
	Version     uint64   `field:"version"`     //
	State       uint8    `field:"state"`       // 状态
}

type NSRecordOperator struct {
	Id          interface{} // ID
	DomainId    interface{} // 域名ID
	IsOn        interface{} // 是否启用
	Description interface{} // 备注
	Name        interface{} // 记录名
	Type        interface{} // 类型
	Value       interface{} // 值
	Ttl         interface{} // TTL（秒）
	Weight      interface{} // 权重
	RouteIds    interface{} // 线路
	CreatedAt   interface{} // 创建时间
	Version     interface{} //
	State       interface{} // 状态
}

func NewNSRecordOperator() *NSRecordOperator {
	return &NSRecordOperator{}
}
