package models

import "github.com/iwind/TeaGo/dbs"

// ClientSystem 终端操作系统信息
type ClientSystem struct {
	Id    uint32   `field:"id"`    // ID
	Name  string   `field:"name"`  // 系统名称
	Codes dbs.JSON `field:"codes"` // 代号
	State uint8    `field:"state"` //
}

type ClientSystemOperator struct {
	Id    interface{} // ID
	Name  interface{} // 系统名称
	Codes interface{} // 代号
	State interface{} //
}

func NewClientSystemOperator() *ClientSystemOperator {
	return &ClientSystemOperator{}
}
