package accounts

// UserAccountLog 用户账户日志
type UserAccountLog struct {
	Id          uint64  `field:"id"`          // ID
	UserId      uint64  `field:"userId"`      // 用户ID
	AccountId   uint64  `field:"accountId"`   // 账户ID
	Delta       float64 `field:"delta"`       // 操作余额的数量（可为负）
	DeltaFrozen float64 `field:"deltaFrozen"` // 操作冻结的数量（可为负）
	Total       float64 `field:"total"`       // 操作后余额
	TotalFrozen float64 `field:"totalFrozen"` // 操作后冻结余额
	EventType   string  `field:"eventType"`   // 类型
	Description string  `field:"description"` // 描述文字
	Day         string  `field:"day"`         // YYYYMMDD
	CreatedAt   uint64  `field:"createdAt"`   // 时间
	Params      string  `field:"params"`      // 参数
}

type UserAccountLogOperator struct {
	Id          interface{} // ID
	UserId      interface{} // 用户ID
	AccountId   interface{} // 账户ID
	Delta       interface{} // 操作余额的数量（可为负）
	DeltaFrozen interface{} // 操作冻结的数量（可为负）
	Total       interface{} // 操作后余额
	TotalFrozen interface{} // 操作后冻结余额
	EventType   interface{} // 类型
	Description interface{} // 描述文字
	Day         interface{} // YYYYMMDD
	CreatedAt   interface{} // 时间
	Params      interface{} // 参数
}

func NewUserAccountLogOperator() *UserAccountLogOperator {
	return &UserAccountLogOperator{}
}
