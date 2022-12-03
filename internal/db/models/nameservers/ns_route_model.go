package nameservers

import "github.com/iwind/TeaGo/dbs"

// NSRoute DNS线路
type NSRoute struct {
	Id         uint32   `field:"id"`         // ID
	IsOn       bool     `field:"isOn"`       // 是否启用
	ClusterId  uint32   `field:"clusterId"`  // 集群ID
	CategoryId uint32   `field:"categoryId"` // 分类ID
	DomainId   uint64   `field:"domainId"`   // 域名ID
	AdminId    uint64   `field:"adminId"`    // 管理员ID
	UserId     uint64   `field:"userId"`     // 用户ID
	IsPublic   bool     `field:"isPublic"`   // 是否公用（管理员创建的线路）
	Name       string   `field:"name"`       // 名称
	Ranges     dbs.JSON `field:"ranges"`     // 范围
	Order      uint32   `field:"order"`      // 排序
	Version    uint64   `field:"version"`    // 版本号
	Priority   uint32   `field:"priority"`   // 优先级，越高越优先
	Code       string   `field:"code"`       // 代号
	State      uint8    `field:"state"`      // 状态
}

type NSRouteOperator struct {
	Id         any // ID
	IsOn       any // 是否启用
	ClusterId  any // 集群ID
	CategoryId any // 分类ID
	DomainId   any // 域名ID
	AdminId    any // 管理员ID
	UserId     any // 用户ID
	IsPublic   any // 是否公用（管理员创建的线路）
	Name       any // 名称
	Ranges     any // 范围
	Order      any // 排序
	Version    any // 版本号
	Priority   any // 优先级，越高越优先
	Code       any // 代号
	State      any // 状态
}

func NewNSRouteOperator() *NSRouteOperator {
	return &NSRouteOperator{}
}
