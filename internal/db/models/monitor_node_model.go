package models

// 监控节点
type MonitorNode struct {
	Id          uint32 `field:"id"`          // ID
	IsOn        uint8  `field:"isOn"`        // 是否启用
	UniqueId    string `field:"uniqueId"`    // 唯一ID
	Secret      string `field:"secret"`      // 密钥
	Name        string `field:"name"`        // 名称
	Description string `field:"description"` // 描述
	Order       uint32 `field:"order"`       // 排序
	State       uint8  `field:"state"`       // 状态
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	Weight      uint32 `field:"weight"`      // 权重
	Status      string `field:"status"`      // 运行状态
}

type MonitorNodeOperator struct {
	Id          interface{} // ID
	IsOn        interface{} // 是否启用
	UniqueId    interface{} // 唯一ID
	Secret      interface{} // 密钥
	Name        interface{} // 名称
	Description interface{} // 描述
	Order       interface{} // 排序
	State       interface{} // 状态
	CreatedAt   interface{} // 创建时间
	AdminId     interface{} // 管理员ID
	Weight      interface{} // 权重
	Status      interface{} // 运行状态
}

func NewMonitorNodeOperator() *MonitorNodeOperator {
	return &MonitorNodeOperator{}
}
