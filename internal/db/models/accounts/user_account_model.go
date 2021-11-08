package accounts

// UserAccount 用户账号
type UserAccount struct {
	Id          uint64  `field:"id"`          // ID
	UserId      uint64  `field:"userId"`      // 用户ID
	Total       float64 `field:"total"`       // 可用总余额
	TotalFrozen float64 `field:"totalFrozen"` // 冻结余额
}

type UserAccountOperator struct {
	Id          interface{} // ID
	UserId      interface{} // 用户ID
	Total       interface{} // 可用总余额
	TotalFrozen interface{} // 冻结余额
}

func NewUserAccountOperator() *UserAccountOperator {
	return &UserAccountOperator{}
}
