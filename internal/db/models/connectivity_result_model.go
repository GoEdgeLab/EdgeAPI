package models

// ConnectivityResult 连通性监控结果
type ConnectivityResult struct {
	Id           uint64  `field:"id"`           // ID
	Type         string  `field:"type"`         // 对象类型
	TargetId     uint32  `field:"targetId"`     // 对象ID
	UpdatedAt    uint64  `field:"updatedAt"`    // 更新时间
	ReportNodeId uint32  `field:"reportNodeId"` // 监控节点ID
	IsOk         uint8   `field:"isOk"`         // 是否可连接
	CostMs       float64 `field:"costMs"`       // 单次连接花费的时间
	Port         uint32  `field:"port"`         // 连接的端口
	Error        string  `field:"error"`        // 产生的错误信息
}

type ConnectivityResultOperator struct {
	Id           interface{} // ID
	Type         interface{} // 对象类型
	TargetId     interface{} // 对象ID
	UpdatedAt    interface{} // 更新时间
	ReportNodeId interface{} // 监控节点ID
	IsOk         interface{} // 是否可连接
	CostMs       interface{} // 单次连接花费的时间
	Port         interface{} // 连接的端口
	Error        interface{} // 产生的错误信息
}

func NewConnectivityResultOperator() *ConnectivityResultOperator {
	return &ConnectivityResultOperator{}
}
