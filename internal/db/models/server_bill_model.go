package models

// ServerBill 服务账单
type ServerBill struct {
	Id                       uint64  `field:"id"`                       // ID
	UserId                   uint32  `field:"userId"`                   // 用户ID
	ServerId                 uint32  `field:"serverId"`                 // 服务ID
	Amount                   float64 `field:"amount"`                   // 金额
	Month                    string  `field:"month"`                    // 月份
	CreatedAt                uint64  `field:"createdAt"`                // 创建时间
	UserPlanId               uint32  `field:"userPlanId"`               // 用户套餐ID
	PlanId                   uint32  `field:"planId"`                   // 套餐ID
	TotalTrafficBytes        uint64  `field:"totalTrafficBytes"`        // 总流量
	BandwidthPercentileBytes uint64  `field:"bandwidthPercentileBytes"` // 带宽百分位字节
	BandwidthPercentile      uint8   `field:"bandwidthPercentile"`      // 带宽百分位
}

type ServerBillOperator struct {
	Id                       interface{} // ID
	UserId                   interface{} // 用户ID
	ServerId                 interface{} // 服务ID
	Amount                   interface{} // 金额
	Month                    interface{} // 月份
	CreatedAt                interface{} // 创建时间
	UserPlanId               interface{} // 用户套餐ID
	PlanId                   interface{} // 套餐ID
	TotalTrafficBytes        interface{} // 总流量
	BandwidthPercentileBytes interface{} // 带宽百分位字节
	BandwidthPercentile      interface{} // 带宽百分位
}

func NewServerBillOperator() *ServerBillOperator {
	return &ServerBillOperator{}
}
