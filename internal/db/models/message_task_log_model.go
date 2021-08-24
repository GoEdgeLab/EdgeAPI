package models

// MessageTaskLog 消息发送日志
type MessageTaskLog struct {
	Id        uint64 `field:"id"`        // ID
	TaskId    uint64 `field:"taskId"`    // 任务ID
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	IsOk      uint8  `field:"isOk"`      // 是否成功
	Error     string `field:"error"`     // 错误信息
	Response  string `field:"response"`  // 响应信息
	Day       string `field:"day"`       // YYYYMMDD
}

type MessageTaskLogOperator struct {
	Id        interface{} // ID
	TaskId    interface{} // 任务ID
	CreatedAt interface{} // 创建时间
	IsOk      interface{} // 是否成功
	Error     interface{} // 错误信息
	Response  interface{} // 响应信息
	Day       interface{} // YYYYMMDD
}

func NewMessageTaskLogOperator() *MessageTaskLogOperator {
	return &MessageTaskLogOperator{}
}
