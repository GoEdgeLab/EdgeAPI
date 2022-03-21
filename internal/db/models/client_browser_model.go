package models

import "github.com/iwind/TeaGo/dbs"

// ClientBrowser 终端浏览器信息
type ClientBrowser struct {
	Id    uint32   `field:"id"`    // ID
	Name  string   `field:"name"`  // 浏览器名称
	Codes dbs.JSON `field:"codes"` // 代号
	State uint8    `field:"state"` // 状态
}

type ClientBrowserOperator struct {
	Id    interface{} // ID
	Name  interface{} // 浏览器名称
	Codes interface{} // 代号
	State interface{} // 状态
}

func NewClientBrowserOperator() *ClientBrowserOperator {
	return &ClientBrowserOperator{}
}
