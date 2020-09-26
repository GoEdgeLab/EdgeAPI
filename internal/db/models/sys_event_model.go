package models

// 系统事件
type SysEvent struct {
	Id        uint64 `field:"id"`        // ID
	Type      string `field:"type"`      // 类型
	Params    string `field:"params"`    // 参数
	CreatedAt uint64 `field:"createdAt"` // 创建时间
}

type SysEventOperator struct {
	Id        interface{} // ID
	Type      interface{} // 类型
	Params    interface{} // 参数
	CreatedAt interface{} // 创建时间
}

func NewSysEventOperator() *SysEventOperator {
	return &SysEventOperator{}
}
