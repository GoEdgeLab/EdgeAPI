package models

import "github.com/iwind/TeaGo/dbs"

// MessageRecipient 消息媒介接收人
type MessageRecipient struct {
	Id          uint32   `field:"id"`          // ID
	AdminId     uint32   `field:"adminId"`     // 管理员ID
	IsOn        uint8    `field:"isOn"`        // 是否启用
	InstanceId  uint32   `field:"instanceId"`  // 实例ID
	User        string   `field:"user"`        // 接收人信息
	GroupIds    dbs.JSON `field:"groupIds"`    // 分组ID
	State       uint8    `field:"state"`       // 状态
	TimeFrom    string   `field:"timeFrom"`    // 开始时间
	TimeTo      string   `field:"timeTo"`      // 结束时间
	Description string   `field:"description"` // 备注
}

type MessageRecipientOperator struct {
	Id          interface{} // ID
	AdminId     interface{} // 管理员ID
	IsOn        interface{} // 是否启用
	InstanceId  interface{} // 实例ID
	User        interface{} // 接收人信息
	GroupIds    interface{} // 分组ID
	State       interface{} // 状态
	TimeFrom    interface{} // 开始时间
	TimeTo      interface{} // 结束时间
	Description interface{} // 备注
}

func NewMessageRecipientOperator() *MessageRecipientOperator {
	return &MessageRecipientOperator{}
}
