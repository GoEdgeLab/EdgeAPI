package models

// UserBill 用户账单
type UserBill struct {
	Id          uint64  `field:"id"`          // ID
	UserId      uint32  `field:"userId"`      // 用户ID
	Type        string  `field:"type"`        // 消费类型
	Description string  `field:"description"` // 描述
	Amount      float64 `field:"amount"`      // 消费数额
	DayFrom     string  `field:"dayFrom"`     // YYYYMMDD
	DayTo       string  `field:"dayTo"`       // YYYYMMDD
	Month       string  `field:"month"`       // 帐期YYYYMM
	CanPay      bool    `field:"canPay"`      // 是否可以支付
	IsPaid      bool    `field:"isPaid"`      // 是否已支付
	PaidAt      uint64  `field:"paidAt"`      // 支付时间
	Code        string  `field:"code"`        // 账单编号
	CreatedAt   uint64  `field:"createdAt"`   // 创建时间
}

type UserBillOperator struct {
	Id          interface{} // ID
	UserId      interface{} // 用户ID
	Type        interface{} // 消费类型
	Description interface{} // 描述
	Amount      interface{} // 消费数额
	DayFrom     interface{} // YYYYMMDD
	DayTo       interface{} // YYYYMMDD
	Month       interface{} // 帐期YYYYMM
	CanPay      interface{} // 是否可以支付
	IsPaid      interface{} // 是否已支付
	PaidAt      interface{} // 支付时间
	Code        interface{} // 账单编号
	CreatedAt   interface{} // 创建时间
}

func NewUserBillOperator() *UserBillOperator {
	return &UserBillOperator{}
}
