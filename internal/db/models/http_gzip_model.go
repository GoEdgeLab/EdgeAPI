package models

import "github.com/iwind/TeaGo/dbs"

// HTTPGzip Gzip配置
type HTTPGzip struct {
	Id        uint32   `field:"id"`        // ID
	AdminId   uint32   `field:"adminId"`   // 管理员ID
	UserId    uint32   `field:"userId"`    // 用户ID
	IsOn      bool     `field:"isOn"`      // 是否启用
	Level     uint32   `field:"level"`     // 压缩级别
	MinLength dbs.JSON `field:"minLength"` // 可压缩最小值
	MaxLength dbs.JSON `field:"maxLength"` // 可压缩最大值
	State     uint8    `field:"state"`     // 状态
	CreatedAt uint64   `field:"createdAt"` // 创建时间
	Conds     dbs.JSON `field:"conds"`     // 条件
}

type HTTPGzipOperator struct {
	Id        interface{} // ID
	AdminId   interface{} // 管理员ID
	UserId    interface{} // 用户ID
	IsOn      interface{} // 是否启用
	Level     interface{} // 压缩级别
	MinLength interface{} // 可压缩最小值
	MaxLength interface{} // 可压缩最大值
	State     interface{} // 状态
	CreatedAt interface{} // 创建时间
	Conds     interface{} // 条件
}

func NewHTTPGzipOperator() *HTTPGzipOperator {
	return &HTTPGzipOperator{}
}
