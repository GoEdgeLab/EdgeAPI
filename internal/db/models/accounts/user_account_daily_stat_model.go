package accounts

// UserAccountDailyStat 账户每日统计
type UserAccountDailyStat struct {
	Id      uint32  `field:"id"`      // ID
	Day     string  `field:"day"`     // YYYYMMDD
	Month   string  `field:"month"`   // YYYYMM
	Income  float64 `field:"income"`  // 收入
	Expense float64 `field:"expense"` // 支出
}

type UserAccountDailyStatOperator struct {
	Id      interface{} // ID
	Day     interface{} // YYYYMMDD
	Month   interface{} // YYYYMM
	Income  interface{} // 收入
	Expense interface{} // 支出
}

func NewUserAccountDailyStatOperator() *UserAccountDailyStatOperator {
	return &UserAccountDailyStatOperator{}
}
