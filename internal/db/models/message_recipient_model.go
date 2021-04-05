package models

// 消息媒介接收人
type MessageRecipient struct {
	Id          uint32 `field:"id"`          // ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	IsOn        uint8  `field:"isOn"`        // 是否启用
	InstanceId  uint32 `field:"instanceId"`  // 实例ID
	User        string `field:"user"`        // 接收人信息
	GroupIds    string `field:"groupIds"`    // 分组ID
	State       uint8  `field:"state"`       // 状态
	Description string `field:"description"` // 备注
}

type MessageRecipientOperator struct {
	Id          interface{} // ID
	AdminId     interface{} // 管理员ID
	IsOn        interface{} // 是否启用
	InstanceId  interface{} // 实例ID
	User        interface{} // 接收人信息
	GroupIds    interface{} // 分组ID
	State       interface{} // 状态
	Description interface{} // 备注
}

func NewMessageRecipientOperator() *MessageRecipientOperator {
	return &MessageRecipientOperator{}
}
