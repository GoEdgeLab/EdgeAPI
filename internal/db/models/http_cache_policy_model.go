package models

import "github.com/iwind/TeaGo/dbs"

// HTTPCachePolicy HTTP缓存策略
type HTTPCachePolicy struct {
	Id                   uint32   `field:"id"`                   // ID
	AdminId              uint32   `field:"adminId"`              // 管理员ID
	UserId               uint32   `field:"userId"`               // 用户ID
	TemplateId           uint32   `field:"templateId"`           // 模版ID
	IsOn                 uint8    `field:"isOn"`                 // 是否启用
	Name                 string   `field:"name"`                 // 名称
	Capacity             dbs.JSON `field:"capacity"`             // 容量数据
	MaxKeys              uint64   `field:"maxKeys"`              // 最多Key值
	MaxSize              dbs.JSON `field:"maxSize"`              // 最大缓存内容尺寸
	Type                 string   `field:"type"`                 // 存储类型
	Options              dbs.JSON `field:"options"`              // 存储选项
	CreatedAt            uint64   `field:"createdAt"`            // 创建时间
	State                uint8    `field:"state"`                // 状态
	Description          string   `field:"description"`          // 描述
	Refs                 dbs.JSON `field:"refs"`                 // 默认的缓存设置
	SyncCompressionCache uint8    `field:"syncCompressionCache"` // 是否同步写入压缩缓存
}

type HTTPCachePolicyOperator struct {
	Id                   interface{} // ID
	AdminId              interface{} // 管理员ID
	UserId               interface{} // 用户ID
	TemplateId           interface{} // 模版ID
	IsOn                 interface{} // 是否启用
	Name                 interface{} // 名称
	Capacity             interface{} // 容量数据
	MaxKeys              interface{} // 最多Key值
	MaxSize              interface{} // 最大缓存内容尺寸
	Type                 interface{} // 存储类型
	Options              interface{} // 存储选项
	CreatedAt            interface{} // 创建时间
	State                interface{} // 状态
	Description          interface{} // 描述
	Refs                 interface{} // 默认的缓存设置
	SyncCompressionCache interface{} // 是否同步写入压缩缓存
}

func NewHTTPCachePolicyOperator() *HTTPCachePolicyOperator {
	return &HTTPCachePolicyOperator{}
}
