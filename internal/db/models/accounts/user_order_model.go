package accounts

import "github.com/iwind/TeaGo/dbs"

// UserOrder 用户订单
type UserOrder struct {
	Id          uint64   `field:"id"`          // 用户订单
	UserId      uint64   `field:"userId"`      // 用户ID
	Code        string   `field:"code"`        // 订单号
	Type        string   `field:"type"`        // 订单类型
	MethodId    uint32   `field:"methodId"`    // 支付方式
	Status      string   `field:"status"`      // 订单状态
	Amount      float64  `field:"amount"`      // 金额
	Params      dbs.JSON `field:"params"`      // 附加参数
	ExpiredAt   uint64   `field:"expiredAt"`   // 过期时间
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
	CancelledAt uint64   `field:"cancelledAt"` // 取消时间
	FinishedAt  uint64   `field:"finishedAt"`  // 结束时间
	State       uint8    `field:"state"`       // 状态
}

type UserOrderOperator struct {
	Id          interface{} // 用户订单
	UserId      interface{} // 用户ID
	Code        interface{} // 订单号
	Type        interface{} // 订单类型
	MethodId    interface{} // 支付方式
	Status      interface{} // 订单状态
	Amount      interface{} // 金额
	Params      interface{} // 附加参数
	ExpiredAt   interface{} // 过期时间
	CreatedAt   interface{} // 创建时间
	CancelledAt interface{} // 取消时间
	FinishedAt  interface{} // 结束时间
	State       interface{} // 状态
}

func NewUserOrderOperator() *UserOrderOperator {
	return &UserOrderOperator{}
}
