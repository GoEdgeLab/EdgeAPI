package models

// 管理员
type Admin struct {
	Id        uint32 `field:"id"`        // ID
	IsOn      uint8  `field:"isOn"`      // 是否启用
	Username  string `field:"username"`  // 用户名
	Password  string `field:"password"`  // 密码
	Fullname  string `field:"fullname"`  // 全名
	IsSuper   uint8  `field:"isSuper"`   // 是否为超级管理员
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	UpdatedAt uint64 `field:"updatedAt"` // 修改时间
	State     uint8  `field:"state"`     // 状态
	Modules   string `field:"modules"`   // 允许的模块
}

type AdminOperator struct {
	Id        interface{} // ID
	IsOn      interface{} // 是否启用
	Username  interface{} // 用户名
	Password  interface{} // 密码
	Fullname  interface{} // 全名
	IsSuper   interface{} // 是否为超级管理员
	CreatedAt interface{} // 创建时间
	UpdatedAt interface{} // 修改时间
	State     interface{} // 状态
	Modules   interface{} // 允许的模块
}

func NewAdminOperator() *AdminOperator {
	return &AdminOperator{}
}
