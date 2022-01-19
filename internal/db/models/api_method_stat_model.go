package models

// APIMethodStat API方法统计
type APIMethodStat struct {
	Id         uint64  `field:"id"`         // ID
	ApiNodeId  uint32  `field:"apiNodeId"`  // API节点ID
	Method     string  `field:"method"`     // 方法
	Tag        string  `field:"tag"`        // 标签方法
	CostMs     float64 `field:"costMs"`     // 耗时Ms
	PeekMs     float64 `field:"peekMs"`     // 峰值耗时
	CountCalls uint64  `field:"countCalls"` // 调用次数
	Day        string  `field:"day"`        // 日期
}

type APIMethodStatOperator struct {
	Id         interface{} // ID
	ApiNodeId  interface{} // API节点ID
	Method     interface{} // 方法
	Tag        interface{} // 标签方法
	CostMs     interface{} // 耗时Ms
	PeekMs     interface{} // 峰值耗时
	CountCalls interface{} // 调用次数
	Day        interface{} // 日期
}

func NewAPIMethodStatOperator() *APIMethodStatOperator {
	return &APIMethodStatOperator{}
}
