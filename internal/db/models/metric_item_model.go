package models

import "github.com/iwind/TeaGo/dbs"

// MetricItem 指标定义
type MetricItem struct {
	Id         uint64   `field:"id"`         // ID
	IsOn       bool     `field:"isOn"`       // 是否启用
	Code       string   `field:"code"`       // 代号（用来区分是否内置）
	Category   string   `field:"category"`   // 类型，比如http, tcp等
	AdminId    uint32   `field:"adminId"`    // 管理员ID
	UserId     uint32   `field:"userId"`     // 用户ID
	Name       string   `field:"name"`       // 指标名称
	Keys       dbs.JSON `field:"keys"`       // 统计的Key
	Period     uint32   `field:"period"`     // 周期
	PeriodUnit string   `field:"periodUnit"` // 周期单位
	Value      string   `field:"value"`      // 值运算
	State      uint8    `field:"state"`      // 状态
	Version    uint32   `field:"version"`    // 版本号
	IsPublic   bool     `field:"isPublic"`   // 是否为公用
}

type MetricItemOperator struct {
	Id         interface{} // ID
	IsOn       interface{} // 是否启用
	Code       interface{} // 代号（用来区分是否内置）
	Category   interface{} // 类型，比如http, tcp等
	AdminId    interface{} // 管理员ID
	UserId     interface{} // 用户ID
	Name       interface{} // 指标名称
	Keys       interface{} // 统计的Key
	Period     interface{} // 周期
	PeriodUnit interface{} // 周期单位
	Value      interface{} // 值运算
	State      interface{} // 状态
	Version    interface{} // 版本号
	IsPublic   interface{} // 是否为公用
}

func NewMetricItemOperator() *MetricItemOperator {
	return &MetricItemOperator{}
}
