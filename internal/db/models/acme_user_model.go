package models

//
type ACMEUser struct {
	Id          uint64 `field:"id"`          // ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	UserId      uint32 `field:"userId"`      // 用户ID
	PrivateKey  string `field:"privateKey"`  // 私钥
	Email       string `field:"email"`       // E-mail
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	State       uint8  `field:"state"`       // 状态
	Description string `field:"description"` // 备注介绍
}

type ACMEUserOperator struct {
	Id          interface{} // ID
	AdminId     interface{} // 管理员ID
	UserId      interface{} // 用户ID
	PrivateKey  interface{} // 私钥
	Email       interface{} // E-mail
	CreatedAt   interface{} // 创建时间
	State       interface{} // 状态
	Description interface{} // 备注介绍
}

func NewACMEUserOperator() *ACMEUserOperator {
	return &ACMEUserOperator{}
}
