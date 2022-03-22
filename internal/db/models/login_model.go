package models

import "github.com/iwind/TeaGo/dbs"

// Login 第三方登录认证
type Login struct {
	Id      uint32   `field:"id"`      // ID
	AdminId uint32   `field:"adminId"` // 管理员ID
	UserId  uint32   `field:"userId"`  // 用户ID
	IsOn    bool     `field:"isOn"`    // 是否启用
	Type    string   `field:"type"`    // 认证方式
	Params  dbs.JSON `field:"params"`  // 参数
	State   uint8    `field:"state"`   // 状态
}

type LoginOperator struct {
	Id      interface{} // ID
	AdminId interface{} // 管理员ID
	UserId  interface{} // 用户ID
	IsOn    interface{} // 是否启用
	Type    interface{} // 认证方式
	Params  interface{} // 参数
	State   interface{} // 状态
}

func NewLoginOperator() *LoginOperator {
	return &LoginOperator{}
}
