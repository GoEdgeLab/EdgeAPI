package models

// MessageMedia 消息媒介
type MessageMedia struct {
	Id              uint32 `field:"id"`              // ID
	Name            string `field:"name"`            // 名称
	Type            string `field:"type"`            // 类型
	Description     string `field:"description"`     // 描述
	UserDescription string `field:"userDescription"` // 用户描述
	IsOn            bool   `field:"isOn"`            // 是否启用
	Order           uint32 `field:"order"`           // 排序
	State           uint8  `field:"state"`           // 状态
}

type MessageMediaOperator struct {
	Id              interface{} // ID
	Name            interface{} // 名称
	Type            interface{} // 类型
	Description     interface{} // 描述
	UserDescription interface{} // 用户描述
	IsOn            interface{} // 是否启用
	Order           interface{} // 排序
	State           interface{} // 状态
}

func NewMessageMediaOperator() *MessageMediaOperator {
	return &MessageMediaOperator{}
}
