package models

// MessageRecipientGroup 消息接收人分组
type MessageRecipientGroup struct {
	Id    uint32 `field:"id"`    // ID
	Name  string `field:"name"`  // 分组名
	Order uint32 `field:"order"` // 排序
	IsOn  bool   `field:"isOn"`  // 是否启用
	State uint8  `field:"state"` // 状态
}

type MessageRecipientGroupOperator struct {
	Id    interface{} // ID
	Name  interface{} // 分组名
	Order interface{} // 排序
	IsOn  interface{} // 是否启用
	State interface{} // 状态
}

func NewMessageRecipientGroupOperator() *MessageRecipientGroupOperator {
	return &MessageRecipientGroupOperator{}
}
