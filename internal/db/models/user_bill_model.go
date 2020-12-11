package models

// 用户账单
type UserBill struct {
	Id          uint64  `field:"id"`          // ID
	UserId      uint32  `field:"userId"`      // 用户ID
	Type        string  `field:"type"`        // 消费类型
	Description string  `field:"description"` // 描述
	Amount      float64 `field:"amount"`      // 消费数额
	Month       string  `field:"month"`       // 帐期YYYYMM
	IsPaid      uint8   `field:"isPaid"`      // 是否已支付
	PaidAt      uint64  `field:"paidAt"`      // 支付时间
	CreatedAt   uint64  `field:"createdAt"`   // 创建时间
}

type UserBillOperator struct {
	Id          interface{} // ID
	UserId      interface{} // 用户ID
	Type        interface{} // 消费类型
	Description interface{} // 描述
	Amount      interface{} // 消费数额
	Month       interface{} // 帐期YYYYMM
	IsPaid      interface{} // 是否已支付
	PaidAt      interface{} // 支付时间
	CreatedAt   interface{} // 创建时间
}

func NewUserBillOperator() *UserBillOperator {
	return &UserBillOperator{}
}
