package stats

// 操作系统统计（按月）
type ServerClientSystemMonthlyStat struct {
	Id       uint64 `field:"id"`       // ID
	ServerId uint32 `field:"serverId"` // 服务ID
	SystemId uint32 `field:"systemId"` // 系统ID
	Version  string `field:"version"`  // 主版本号
	Month    string `field:"month"`    // YYYYMM
	Count    uint64 `field:"count"`    // 数量
}

type ServerClientSystemMonthlyStatOperator struct {
	Id       interface{} // ID
	ServerId interface{} // 服务ID
	SystemId interface{} // 系统ID
	Version  interface{} // 主版本号
	Month    interface{} // YYYYMM
	Count    interface{} // 数量
}

func NewServerClientSystemMonthlyStatOperator() *ServerClientSystemMonthlyStatOperator {
	return &ServerClientSystemMonthlyStatOperator{}
}
