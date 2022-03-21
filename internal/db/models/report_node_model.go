package models

import "github.com/iwind/TeaGo/dbs"

// ReportNode 连通性报告终端
type ReportNode struct {
	Id        uint32   `field:"id"`        // ID
	UniqueId  string   `field:"uniqueId"`  // 唯一ID
	Secret    string   `field:"secret"`    // 密钥
	IsOn      uint8    `field:"isOn"`      // 是否启用
	Name      string   `field:"name"`      // 名称
	Location  string   `field:"location"`  // 所在区域
	Isp       string   `field:"isp"`       // 网络服务商
	AllowIPs  dbs.JSON `field:"allowIPs"`  // 允许的IP
	IsActive  uint8    `field:"isActive"`  // 是否活跃
	Status    dbs.JSON `field:"status"`    // 状态
	State     uint8    `field:"state"`     // 状态
	CreatedAt uint64   `field:"createdAt"` // 创建时间
	GroupIds  dbs.JSON `field:"groupIds"`  // 分组ID
}

type ReportNodeOperator struct {
	Id        interface{} // ID
	UniqueId  interface{} // 唯一ID
	Secret    interface{} // 密钥
	IsOn      interface{} // 是否启用
	Name      interface{} // 名称
	Location  interface{} // 所在区域
	Isp       interface{} // 网络服务商
	AllowIPs  interface{} // 允许的IP
	IsActive  interface{} // 是否活跃
	Status    interface{} // 状态
	State     interface{} // 状态
	CreatedAt interface{} // 创建时间
	GroupIds  interface{} // 分组ID
}

func NewReportNodeOperator() *ReportNodeOperator {
	return &ReportNodeOperator{}
}
