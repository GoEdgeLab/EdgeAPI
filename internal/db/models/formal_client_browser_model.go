package models

import "github.com/iwind/TeaGo/dbs"

// FormalClientBrowser 终端浏览器信息
type FormalClientBrowser struct {
	Id     uint32   `field:"id"`     // ID
	Name   string   `field:"name"`   // 浏览器名称
	Codes  dbs.JSON `field:"codes"`  // 代号
	DataId string   `field:"dataId"` // 数据ID
	State  uint8    `field:"state"`  // 状态
}

type FormalClientBrowserOperator struct {
	Id     any // ID
	Name   any // 浏览器名称
	Codes  any // 代号
	DataId any // 数据ID
	State  any // 状态
}

func NewFormalClientBrowserOperator() *FormalClientBrowserOperator {
	return &FormalClientBrowserOperator{}
}
