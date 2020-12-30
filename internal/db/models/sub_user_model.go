package models

// 子用户
type SubUser struct {
	Id       uint32 `field:"id"`       // ID
	UserId   uint32 `field:"userId"`   // 所属主用户ID
	IsOn     uint8  `field:"isOn"`     // 是否启用
	Name     string `field:"name"`     // 名称
	Username string `field:"username"` // 用户名
	Password string `field:"password"` // 密码
	State    uint8  `field:"state"`    // 状态
}

type SubUserOperator struct {
	Id       interface{} // ID
	UserId   interface{} // 所属主用户ID
	IsOn     interface{} // 是否启用
	Name     interface{} // 名称
	Username interface{} // 用户名
	Password interface{} // 密码
	State    interface{} // 状态
}

func NewSubUserOperator() *SubUserOperator {
	return &SubUserOperator{}
}
