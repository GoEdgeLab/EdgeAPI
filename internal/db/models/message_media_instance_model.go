package models

// MessageMediaInstance 消息媒介接收人
type MessageMediaInstance struct {
	Id          uint32 `field:"id"`          // ID
	Name        string `field:"name"`        // 名称
	IsOn        uint8  `field:"isOn"`        // 是否启用
	MediaType   string `field:"mediaType"`   // 媒介类型
	Params      string `field:"params"`      // 媒介参数
	Description string `field:"description"` // 备注
	Rate        string `field:"rate"`        // 发送频率
	State       uint8  `field:"state"`       // 状态
}

type MessageMediaInstanceOperator struct {
	Id          interface{} // ID
	Name        interface{} // 名称
	IsOn        interface{} // 是否启用
	MediaType   interface{} // 媒介类型
	Params      interface{} // 媒介参数
	Description interface{} // 备注
	Rate        interface{} // 发送频率
	State       interface{} // 状态
}

func NewMessageMediaInstanceOperator() *MessageMediaInstanceOperator {
	return &MessageMediaInstanceOperator{}
}
