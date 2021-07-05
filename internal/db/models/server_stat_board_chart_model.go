package models

// ServerStatBoardChart 服务看板中的图表
type ServerStatBoardChart struct {
	Id      uint64 `field:"id"`      // ID
	BoardId uint64 `field:"boardId"` // 看板ID
	Code    string `field:"code"`    // 内置图表代码
	ItemId  uint32 `field:"itemId"`  // 指标ID
	ChartId uint32 `field:"chartId"` // 图表ID
	Order   uint32 `field:"order"`   // 排序
	State   uint8  `field:"state"`   // 状态
}

type ServerStatBoardChartOperator struct {
	Id      interface{} // ID
	BoardId interface{} // 看板ID
	Code    interface{} // 内置图表代码
	ItemId  interface{} // 指标ID
	ChartId interface{} // 图表ID
	Order   interface{} // 排序
	State   interface{} // 状态
}

func NewServerStatBoardChartOperator() *ServerStatBoardChartOperator {
	return &ServerStatBoardChartOperator{}
}
