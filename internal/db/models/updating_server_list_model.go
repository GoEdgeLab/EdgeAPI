package models

import "github.com/iwind/TeaGo/dbs"

// UpdatingServerList 待更新服务列表
type UpdatingServerList struct {
	Id        uint64   `field:"id"`        // ID
	ClusterId uint32   `field:"clusterId"` // 集群ID
	UniqueId  string   `field:"uniqueId"`  // 唯一ID
	ServerIds dbs.JSON `field:"serverIds"` // 服务IDs
	Day       string   `field:"day"`       // 创建日期
}

type UpdatingServerListOperator struct {
	Id        any // ID
	ClusterId any // 集群ID
	UniqueId  any // 唯一ID
	ServerIds any // 服务IDs
	Day       any // 创建日期
}

func NewUpdatingServerListOperator() *UpdatingServerListOperator {
	return &UpdatingServerListOperator{}
}
