package models

// ReportResult 连通性监控结果
type ReportResult struct {
	Id           uint64  `field:"id"`           // ID
	Type         string  `field:"type"`         // 对象类型
	TargetId     uint64  `field:"targetId"`     // 对象ID
	TargetDesc   string  `field:"targetDesc"`   // 对象描述
	UpdatedAt    uint64  `field:"updatedAt"`    // 更新时间
	ReportNodeId uint32  `field:"reportNodeId"` // 监控节点ID
	IsOk         bool    `field:"isOk"`         // 是否可连接
	Level        string  `field:"level"`        // 级别
	CostMs       float64 `field:"costMs"`       // 单次连接花费的时间
	Error        string  `field:"error"`        // 产生的错误信息
	CountUp      uint32  `field:"countUp"`      // 连续上线次数
	CountDown    uint32  `field:"countDown"`    // 连续下线次数
}

type ReportResultOperator struct {
	Id           interface{} // ID
	Type         interface{} // 对象类型
	TargetId     interface{} // 对象ID
	TargetDesc   interface{} // 对象描述
	UpdatedAt    interface{} // 更新时间
	ReportNodeId interface{} // 监控节点ID
	IsOk         interface{} // 是否可连接
	Level        interface{} // 级别
	CostMs       interface{} // 单次连接花费的时间
	Error        interface{} // 产生的错误信息
	CountUp      interface{} // 连续上线次数
	CountDown    interface{} // 连续下线次数
}

func NewReportResultOperator() *ReportResultOperator {
	return &ReportResultOperator{}
}
