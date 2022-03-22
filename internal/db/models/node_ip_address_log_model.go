package models

// NodeIPAddressLog IP状态变更日志
type NodeIPAddressLog struct {
	Id          uint64 `field:"id"`          // ID
	AddressId   uint64 `field:"addressId"`   // 地址ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	Description string `field:"description"` // 描述
	CreatedAt   uint64 `field:"createdAt"`   // 操作时间
	IsUp        uint8  `field:"isUp"`        // 是否在线
	IsOn        bool   `field:"isOn"`        // 是否启用
	CanAccess   uint8  `field:"canAccess"`   // 是否可访问
	Day         string `field:"day"`         // YYYYMMDD，用来清理
	BackupIP    string `field:"backupIP"`    // 备用IP
}

type NodeIPAddressLogOperator struct {
	Id          interface{} // ID
	AddressId   interface{} // 地址ID
	AdminId     interface{} // 管理员ID
	Description interface{} // 描述
	CreatedAt   interface{} // 操作时间
	IsUp        interface{} // 是否在线
	IsOn        interface{} // 是否启用
	CanAccess   interface{} // 是否可访问
	Day         interface{} // YYYYMMDD，用来清理
	BackupIP    interface{} // 备用IP
}

func NewNodeIPAddressLogOperator() *NodeIPAddressLogOperator {
	return &NodeIPAddressLogOperator{}
}
