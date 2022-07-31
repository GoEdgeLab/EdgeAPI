package accounts

import "github.com/iwind/TeaGo/dbs"

// UserOrderLog 订单日志
type UserOrderLog struct {
	Id        uint64   `field:"id"`        // ID
	AdminId   uint64   `field:"adminId"`   // 管理员ID
	UserId    uint64   `field:"userId"`    // 用户ID
	OrderId   uint64   `field:"orderId"`   // 订单ID
	Status    string   `field:"status"`    // 状态
	Snapshot  dbs.JSON `field:"snapshot"`  // 状态快照
	CreatedAt uint64   `field:"createdAt"` // 创建时间
}

type UserOrderLogOperator struct {
	Id        interface{} // ID
	AdminId   interface{} // 管理员ID
	UserId    interface{} // 用户ID
	OrderId   interface{} // 订单ID
	Status    interface{} // 状态
	Snapshot  interface{} // 状态快照
	CreatedAt interface{} // 创建时间
}

func NewUserOrderLogOperator() *UserOrderLogOperator {
	return &UserOrderLogOperator{}
}
