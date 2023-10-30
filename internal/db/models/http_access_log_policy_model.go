package models

import "github.com/iwind/TeaGo/dbs"

const (
	HTTPAccessLogPolicyField_Id               dbs.FieldName = "id"               // ID
	HTTPAccessLogPolicyField_TemplateId       dbs.FieldName = "templateId"       // 模版ID
	HTTPAccessLogPolicyField_AdminId          dbs.FieldName = "adminId"          // 管理员ID
	HTTPAccessLogPolicyField_UserId           dbs.FieldName = "userId"           // 用户ID
	HTTPAccessLogPolicyField_State            dbs.FieldName = "state"            // 状态
	HTTPAccessLogPolicyField_CreatedAt        dbs.FieldName = "createdAt"        // 创建时间
	HTTPAccessLogPolicyField_Name             dbs.FieldName = "name"             // 名称
	HTTPAccessLogPolicyField_IsOn             dbs.FieldName = "isOn"             // 是否启用
	HTTPAccessLogPolicyField_Type             dbs.FieldName = "type"             // 存储类型
	HTTPAccessLogPolicyField_Options          dbs.FieldName = "options"          // 存储选项
	HTTPAccessLogPolicyField_Conds            dbs.FieldName = "conds"            // 请求条件
	HTTPAccessLogPolicyField_IsPublic         dbs.FieldName = "isPublic"         // 是否为公用
	HTTPAccessLogPolicyField_FirewallOnly     dbs.FieldName = "firewallOnly"     // 是否只记录防火墙相关
	HTTPAccessLogPolicyField_Version          dbs.FieldName = "version"          // 版本号
	HTTPAccessLogPolicyField_DisableDefaultDB dbs.FieldName = "disableDefaultDB" // 是否停止默认数据库存储
)

// HTTPAccessLogPolicy 访问日志策略
type HTTPAccessLogPolicy struct {
	Id               uint32   `field:"id"`               // ID
	TemplateId       uint32   `field:"templateId"`       // 模版ID
	AdminId          uint32   `field:"adminId"`          // 管理员ID
	UserId           uint32   `field:"userId"`           // 用户ID
	State            uint8    `field:"state"`            // 状态
	CreatedAt        uint64   `field:"createdAt"`        // 创建时间
	Name             string   `field:"name"`             // 名称
	IsOn             bool     `field:"isOn"`             // 是否启用
	Type             string   `field:"type"`             // 存储类型
	Options          dbs.JSON `field:"options"`          // 存储选项
	Conds            dbs.JSON `field:"conds"`            // 请求条件
	IsPublic         bool     `field:"isPublic"`         // 是否为公用
	FirewallOnly     uint8    `field:"firewallOnly"`     // 是否只记录防火墙相关
	Version          uint32   `field:"version"`          // 版本号
	DisableDefaultDB bool     `field:"disableDefaultDB"` // 是否停止默认数据库存储
}

type HTTPAccessLogPolicyOperator struct {
	Id               any // ID
	TemplateId       any // 模版ID
	AdminId          any // 管理员ID
	UserId           any // 用户ID
	State            any // 状态
	CreatedAt        any // 创建时间
	Name             any // 名称
	IsOn             any // 是否启用
	Type             any // 存储类型
	Options          any // 存储选项
	Conds            any // 请求条件
	IsPublic         any // 是否为公用
	FirewallOnly     any // 是否只记录防火墙相关
	Version          any // 版本号
	DisableDefaultDB any // 是否停止默认数据库存储
}

func NewHTTPAccessLogPolicyOperator() *HTTPAccessLogPolicyOperator {
	return &HTTPAccessLogPolicyOperator{}
}
