package models

// NodeIPAddressLog IP状态变更日志
type NodeIPAddressLog struct {
	Id          uint64 `field:"id"`          // ID
	AddressId   uint64 `field:"addressId"`   // 地址ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	Description string `field:"description"` // 描述
	CreatedAt   uint64 `field:"createdAt"`   // 操作时间
	Day         string `field:"day"`         // YYYYMMDD，用来清理
}

type NodeIPAddressLogOperator struct {
	Id          interface{} // ID
	AddressId   interface{} // 地址ID
	AdminId     interface{} // 管理员ID
	Description interface{} // 描述
	CreatedAt   interface{} // 操作时间
	Day         interface{} // YYYYMMDD，用来清理
}

func NewNodeIPAddressLogOperator() *NodeIPAddressLogOperator {
	return &NodeIPAddressLogOperator{}
}
