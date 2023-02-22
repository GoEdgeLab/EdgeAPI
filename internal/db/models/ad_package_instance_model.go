package models

import "github.com/iwind/TeaGo/dbs"

// ADPackageInstance 高防实例
type ADPackageInstance struct {
	Id             uint32   `field:"id"`             // ID
	IsOn           bool     `field:"isOn"`           // 是否启用
	PackageId      uint32   `field:"packageId"`      // 规格ID
	ClusterId      uint32   `field:"clusterId"`      // 集群ID
	NodeIds        dbs.JSON `field:"nodeIds"`        // 节点ID
	IpAddresses    dbs.JSON `field:"ipAddresses"`    // IP地址
	UserId         uint64   `field:"userId"`         // 用户ID
	UserDayTo      string   `field:"userDayTo"`      // 用户有效期YYYYMMDD
	UserInstanceId uint64   `field:"userInstanceId"` // 用户实例ID
	State          uint8    `field:"state"`          // 状态
	ObjectCodes    dbs.JSON `field:"objectCodes"`    // 防护对象
}

type ADPackageInstanceOperator struct {
	Id             any // ID
	IsOn           any // 是否启用
	PackageId      any // 规格ID
	ClusterId      any // 集群ID
	NodeIds        any // 节点ID
	IpAddresses    any // IP地址
	UserId         any // 用户ID
	UserDayTo      any // 用户有效期YYYYMMDD
	UserInstanceId any // 用户实例ID
	State          any // 状态
	ObjectCodes    any // 防护对象
}

func NewADPackageInstanceOperator() *ADPackageInstanceOperator {
	return &ADPackageInstanceOperator{}
}
