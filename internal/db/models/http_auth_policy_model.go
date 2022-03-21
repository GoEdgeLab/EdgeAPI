package models

import "github.com/iwind/TeaGo/dbs"

// HTTPAuthPolicy HTTP认证策略
type HTTPAuthPolicy struct {
	Id      uint64   `field:"id"`      // ID
	AdminId uint32   `field:"adminId"` // 管理员ID
	UserId  uint32   `field:"userId"`  // 用户ID
	IsOn    uint8    `field:"isOn"`    // 是否启用
	Name    string   `field:"name"`    // 名称
	Type    string   `field:"type"`    // 类型
	Params  dbs.JSON `field:"params"`  // 参数
	State   uint8    `field:"state"`   // 状态
}

type HTTPAuthPolicyOperator struct {
	Id      interface{} // ID
	AdminId interface{} // 管理员ID
	UserId  interface{} // 用户ID
	IsOn    interface{} // 是否启用
	Name    interface{} // 名称
	Type    interface{} // 类型
	Params  interface{} // 参数
	State   interface{} // 状态
}

func NewHTTPAuthPolicyOperator() *HTTPAuthPolicyOperator {
	return &HTTPAuthPolicyOperator{}
}
