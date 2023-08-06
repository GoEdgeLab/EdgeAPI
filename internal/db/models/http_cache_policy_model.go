package models

import "github.com/iwind/TeaGo/dbs"

const (
	HTTPCachePolicyField_Id                   dbs.FieldName = "id"                   // ID
	HTTPCachePolicyField_AdminId              dbs.FieldName = "adminId"              // 管理员ID
	HTTPCachePolicyField_UserId               dbs.FieldName = "userId"               // 用户ID
	HTTPCachePolicyField_TemplateId           dbs.FieldName = "templateId"           // 模版ID
	HTTPCachePolicyField_IsOn                 dbs.FieldName = "isOn"                 // 是否启用
	HTTPCachePolicyField_Name                 dbs.FieldName = "name"                 // 名称
	HTTPCachePolicyField_Capacity             dbs.FieldName = "capacity"             // 容量数据
	HTTPCachePolicyField_MaxKeys              dbs.FieldName = "maxKeys"              // 最多Key值
	HTTPCachePolicyField_MaxSize              dbs.FieldName = "maxSize"              // 最大缓存内容尺寸
	HTTPCachePolicyField_Type                 dbs.FieldName = "type"                 // 存储类型
	HTTPCachePolicyField_Options              dbs.FieldName = "options"              // 存储选项
	HTTPCachePolicyField_CreatedAt            dbs.FieldName = "createdAt"            // 创建时间
	HTTPCachePolicyField_State                dbs.FieldName = "state"                // 状态
	HTTPCachePolicyField_Description          dbs.FieldName = "description"          // 描述
	HTTPCachePolicyField_Refs                 dbs.FieldName = "refs"                 // 默认的缓存设置
	HTTPCachePolicyField_SyncCompressionCache dbs.FieldName = "syncCompressionCache" // 是否同步写入压缩缓存
	HTTPCachePolicyField_FetchTimeout         dbs.FieldName = "fetchTimeout"         // 预热超时时间
)

// HTTPCachePolicy HTTP缓存策略
type HTTPCachePolicy struct {
	Id                   uint32   `field:"id"`                   // ID
	AdminId              uint32   `field:"adminId"`              // 管理员ID
	UserId               uint32   `field:"userId"`               // 用户ID
	TemplateId           uint32   `field:"templateId"`           // 模版ID
	IsOn                 bool     `field:"isOn"`                 // 是否启用
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
	FetchTimeout         dbs.JSON `field:"fetchTimeout"`         // 预热超时时间
}

type HTTPCachePolicyOperator struct {
	Id                   any // ID
	AdminId              any // 管理员ID
	UserId               any // 用户ID
	TemplateId           any // 模版ID
	IsOn                 any // 是否启用
	Name                 any // 名称
	Capacity             any // 容量数据
	MaxKeys              any // 最多Key值
	MaxSize              any // 最大缓存内容尺寸
	Type                 any // 存储类型
	Options              any // 存储选项
	CreatedAt            any // 创建时间
	State                any // 状态
	Description          any // 描述
	Refs                 any // 默认的缓存设置
	SyncCompressionCache any // 是否同步写入压缩缓存
	FetchTimeout         any // 预热超时时间
}

func NewHTTPCachePolicyOperator() *HTTPCachePolicyOperator {
	return &HTTPCachePolicyOperator{}
}
