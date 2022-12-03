package models

import "github.com/iwind/TeaGo/dbs"

// ClientSystem 终端操作系统信息
type ClientSystem struct {
	Id         uint64   `field:"id"`         // ID
	Name       string   `field:"name"`       // 系统名称
	Codes      dbs.JSON `field:"codes"`      // 代号
	CreatedDay string   `field:"createdDay"` // 创建日期YYYYMMDD
	State      uint8    `field:"state"`      // 状态
}

type ClientSystemOperator struct {
	Id         any // ID
	Name       any // 系统名称
	Codes      any // 代号
	CreatedDay any // 创建日期YYYYMMDD
	State      any // 状态
}

func NewClientSystemOperator() *ClientSystemOperator {
	return &ClientSystemOperator{}
}
