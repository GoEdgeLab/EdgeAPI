package acme

// ACMETaskLog ACME任务运行日志
type ACMETaskLog struct {
	Id        uint64 `field:"id"`        // ID
	TaskId    uint64 `field:"taskId"`    // 任务ID
	IsOk      bool   `field:"isOk"`      // 是否成功
	Error     string `field:"error"`     // 错误信息
	CreatedAt uint64 `field:"createdAt"` // 运行时间
}

type ACMETaskLogOperator struct {
	Id        interface{} // ID
	TaskId    interface{} // 任务ID
	IsOk      interface{} // 是否成功
	Error     interface{} // 错误信息
	CreatedAt interface{} // 运行时间
}

func NewACMETaskLogOperator() *ACMETaskLogOperator {
	return &ACMETaskLogOperator{}
}
