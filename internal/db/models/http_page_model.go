package models

import "github.com/iwind/TeaGo/dbs"

const (
	HTTPPageField_Id                dbs.FieldName = "id"                // ID
	HTTPPageField_AdminId           dbs.FieldName = "adminId"           // 管理员ID
	HTTPPageField_UserId            dbs.FieldName = "userId"            // 用户ID
	HTTPPageField_IsOn              dbs.FieldName = "isOn"              // 是否启用
	HTTPPageField_StatusList        dbs.FieldName = "statusList"        // 状态列表
	HTTPPageField_Url               dbs.FieldName = "url"               // 页面URL
	HTTPPageField_NewStatus         dbs.FieldName = "newStatus"         // 新状态码
	HTTPPageField_State             dbs.FieldName = "state"             // 状态
	HTTPPageField_CreatedAt         dbs.FieldName = "createdAt"         // 创建时间
	HTTPPageField_Body              dbs.FieldName = "body"              // 页面内容
	HTTPPageField_BodyType          dbs.FieldName = "bodyType"          // 内容类型
	HTTPPageField_ExceptURLPatterns dbs.FieldName = "exceptURLPatterns" // 例外URL
	HTTPPageField_OnlyURLPatterns   dbs.FieldName = "onlyURLPatterns"   // 限制URL
)

// HTTPPage 特殊页面
type HTTPPage struct {
	Id                uint32   `field:"id"`                // ID
	AdminId           uint32   `field:"adminId"`           // 管理员ID
	UserId            uint32   `field:"userId"`            // 用户ID
	IsOn              bool     `field:"isOn"`              // 是否启用
	StatusList        dbs.JSON `field:"statusList"`        // 状态列表
	Url               string   `field:"url"`               // 页面URL
	NewStatus         int32    `field:"newStatus"`         // 新状态码
	State             uint8    `field:"state"`             // 状态
	CreatedAt         uint64   `field:"createdAt"`         // 创建时间
	Body              string   `field:"body"`              // 页面内容
	BodyType          string   `field:"bodyType"`          // 内容类型
	ExceptURLPatterns dbs.JSON `field:"exceptURLPatterns"` // 例外URL
	OnlyURLPatterns   dbs.JSON `field:"onlyURLPatterns"`   // 限制URL
}

type HTTPPageOperator struct {
	Id                any // ID
	AdminId           any // 管理员ID
	UserId            any // 用户ID
	IsOn              any // 是否启用
	StatusList        any // 状态列表
	Url               any // 页面URL
	NewStatus         any // 新状态码
	State             any // 状态
	CreatedAt         any // 创建时间
	Body              any // 页面内容
	BodyType          any // 内容类型
	ExceptURLPatterns any // 例外URL
	OnlyURLPatterns   any // 限制URL
}

func NewHTTPPageOperator() *HTTPPageOperator {
	return &HTTPPageOperator{}
}
