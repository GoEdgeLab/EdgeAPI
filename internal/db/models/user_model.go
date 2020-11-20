package models

// 用户
type User struct {
	Id        uint32 `field:"id"`        // ID
	Username  string `field:"username"`  // 用户名
	Password  string `field:"password"`  // 密码
	Fullname  string `field:"fullname"`  // 真实姓名
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	UpdatedAt uint64 `field:"updatedAt"` // 修改时间
	State     uint8  `field:"state"`     // 状态
	Mobile    string `field:"mobile"`    // 手机号
	Remark    string `field:"remark"`    // 备注
	Email     string `field:"email"`     // 邮箱地址
}

type UserOperator struct {
	Id        interface{} // ID
	Username  interface{} // 用户名
	Password  interface{} // 密码
	Fullname  interface{} // 真实姓名
	CreatedAt interface{} // 创建时间
	UpdatedAt interface{} // 修改时间
	State     interface{} // 状态
	Mobile    interface{} // 手机号
	Remark    interface{} // 备注
	Email     interface{} // 邮箱地址
}

func NewUserOperator() *UserOperator {
	return &UserOperator{}
}
