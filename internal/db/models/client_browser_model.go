package models

import "github.com/iwind/TeaGo/dbs"

// ClientBrowser 终端浏览器信息
type ClientBrowser struct {
	Id         uint32   `field:"id"`         // ID
	Name       string   `field:"name"`       // 浏览器名称
	Codes      dbs.JSON `field:"codes"`      // 代号
	CreatedDay string   `field:"createdDay"` // 创建日期YYYYMMDD
	State      uint8    `field:"state"`      // 状态
}

type ClientBrowserOperator struct {
	Id         any // ID
	Name       any // 浏览器名称
	Codes      any // 代号
	CreatedDay any // 创建日期YYYYMMDD
	State      any // 状态
}

func NewClientBrowserOperator() *ClientBrowserOperator {
	return &ClientBrowserOperator{}
}
