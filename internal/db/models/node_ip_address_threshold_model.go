package models

import "github.com/iwind/TeaGo/dbs"

// NodeIPAddressThreshold IP地址阈值
type NodeIPAddressThreshold struct {
	Id         uint64   `field:"id"`         // ID
	AddressId  uint64   `field:"addressId"`  // IP地址ID
	Items      dbs.JSON `field:"items"`      // 阈值条目
	Actions    dbs.JSON `field:"actions"`    // 动作
	NotifiedAt uint64   `field:"notifiedAt"` // 上次通知时间
	IsMatched  uint8    `field:"isMatched"`  // 上次是否匹配
	State      uint8    `field:"state"`      // 状态
	Order      uint32   `field:"order"`      // 排序
}

type NodeIPAddressThresholdOperator struct {
	Id         interface{} // ID
	AddressId  interface{} // IP地址ID
	Items      interface{} // 阈值条目
	Actions    interface{} // 动作
	NotifiedAt interface{} // 上次通知时间
	IsMatched  interface{} // 上次是否匹配
	State      interface{} // 状态
	Order      interface{} // 排序
}

func NewNodeIPAddressThresholdOperator() *NodeIPAddressThresholdOperator {
	return &NodeIPAddressThresholdOperator{}
}
