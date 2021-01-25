package stats

// 浏览器统计（按月）
type ServerClientBrowserMonthlyStat struct {
	Id        uint64 `field:"id"`        // ID
	ServerId  uint32 `field:"serverId"`  // 服务ID
	BrowserId uint32 `field:"browserId"` // 浏览器ID
	Month     string `field:"month"`     // YYYYMM
	Version   string `field:"version"`   // 主版本号
	Count     uint64 `field:"count"`     // 数量
}

type ServerClientBrowserMonthlyStatOperator struct {
	Id        interface{} // ID
	ServerId  interface{} // 服务ID
	BrowserId interface{} // 浏览器ID
	Month     interface{} // YYYYMM
	Version   interface{} // 主版本号
	Count     interface{} // 数量
}

func NewServerClientBrowserMonthlyStatOperator() *ServerClientBrowserMonthlyStatOperator {
	return &ServerClientBrowserMonthlyStatOperator{}
}
