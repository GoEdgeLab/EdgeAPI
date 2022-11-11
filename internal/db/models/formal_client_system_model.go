package models

import "github.com/iwind/TeaGo/dbs"

// FormalClientSystem 终端操作系统信息
type FormalClientSystem struct {
	Id     uint32   `field:"id"`     // ID
	Name   string   `field:"name"`   // 系统名称
	Codes  dbs.JSON `field:"codes"`  // 代号
	State  uint8    `field:"state"`  // 状态
	DataId string   `field:"dataId"` // 数据ID
}

type FormalClientSystemOperator struct {
	Id     any // ID
	Name   any // 系统名称
	Codes  any // 代号
	State  any // 状态
	DataId any // 数据ID
}

func NewFormalClientSystemOperator() *FormalClientSystemOperator {
	return &FormalClientSystemOperator{}
}
