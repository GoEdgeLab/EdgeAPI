package models

// UserBill 用户账单
type UserBill struct {
	Id          uint64  `field:"id"`          // ID
	UserId      uint32  `field:"userId"`      // 用户ID
	Type        string  `field:"type"`        // 消费类型
	PricePeriod string  `field:"pricePeriod"` // 计费周期
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
	CreatedDay  string  `field:"createdDay"`  // 创建日期
	State       uint8   `field:"state"`       // 状态
}

type UserBillOperator struct {
	Id          any // ID
	UserId      any // 用户ID
	Type        any // 消费类型
	PricePeriod any // 计费周期
	Description any // 描述
	Amount      any // 消费数额
	DayFrom     any // YYYYMMDD
	DayTo       any // YYYYMMDD
	Month       any // 帐期YYYYMM
	CanPay      any // 是否可以支付
	IsPaid      any // 是否已支付
	PaidAt      any // 支付时间
	Code        any // 账单编号
	CreatedAt   any // 创建时间
	CreatedDay  any // 创建日期
	State       any // 状态
}

func NewUserBillOperator() *UserBillOperator {
	return &UserBillOperator{}
}
