package models

import "github.com/iwind/TeaGo/dbs"

const (
	HTTPFirewallPolicyField_Id                 dbs.FieldName = "id"                 // ID
	HTTPFirewallPolicyField_TemplateId         dbs.FieldName = "templateId"         // 模版ID
	HTTPFirewallPolicyField_AdminId            dbs.FieldName = "adminId"            // 管理员ID
	HTTPFirewallPolicyField_UserId             dbs.FieldName = "userId"             // 用户ID
	HTTPFirewallPolicyField_ServerId           dbs.FieldName = "serverId"           // 服务ID
	HTTPFirewallPolicyField_GroupId            dbs.FieldName = "groupId"            // 服务分组ID
	HTTPFirewallPolicyField_State              dbs.FieldName = "state"              // 状态
	HTTPFirewallPolicyField_CreatedAt          dbs.FieldName = "createdAt"          // 创建时间
	HTTPFirewallPolicyField_IsOn               dbs.FieldName = "isOn"               // 是否启用
	HTTPFirewallPolicyField_Name               dbs.FieldName = "name"               // 名称
	HTTPFirewallPolicyField_Description        dbs.FieldName = "description"        // 描述
	HTTPFirewallPolicyField_Inbound            dbs.FieldName = "inbound"            // 入站规则
	HTTPFirewallPolicyField_Outbound           dbs.FieldName = "outbound"           // 出站规则
	HTTPFirewallPolicyField_BlockOptions       dbs.FieldName = "blockOptions"       // BLOCK动作选项
	HTTPFirewallPolicyField_PageOptions        dbs.FieldName = "pageOptions"        // PAGE动作选项
	HTTPFirewallPolicyField_CaptchaOptions     dbs.FieldName = "captchaOptions"     // 验证码动作选项
	HTTPFirewallPolicyField_Mode               dbs.FieldName = "mode"               // 模式
	HTTPFirewallPolicyField_UseLocalFirewall   dbs.FieldName = "useLocalFirewall"   // 是否自动使用本地防火墙
	HTTPFirewallPolicyField_SynFlood           dbs.FieldName = "synFlood"           // SynFlood防御设置
	HTTPFirewallPolicyField_Log                dbs.FieldName = "log"                // 日志配置
	HTTPFirewallPolicyField_MaxRequestBodySize dbs.FieldName = "maxRequestBodySize" // 可以检查的最大请求内容尺寸
	HTTPFirewallPolicyField_DenyCountryHTML    dbs.FieldName = "denyCountryHTML"    // 区域封禁提示
	HTTPFirewallPolicyField_DenyProvinceHTML   dbs.FieldName = "denyProvinceHTML"   // 省份封禁提示
)

// HTTPFirewallPolicy HTTP防火墙
type HTTPFirewallPolicy struct {
	Id                 uint32   `field:"id"`                 // ID
	TemplateId         uint32   `field:"templateId"`         // 模版ID
	AdminId            uint32   `field:"adminId"`            // 管理员ID
	UserId             uint32   `field:"userId"`             // 用户ID
	ServerId           uint32   `field:"serverId"`           // 服务ID
	GroupId            uint32   `field:"groupId"`            // 服务分组ID
	State              uint8    `field:"state"`              // 状态
	CreatedAt          uint64   `field:"createdAt"`          // 创建时间
	IsOn               bool     `field:"isOn"`               // 是否启用
	Name               string   `field:"name"`               // 名称
	Description        string   `field:"description"`        // 描述
	Inbound            dbs.JSON `field:"inbound"`            // 入站规则
	Outbound           dbs.JSON `field:"outbound"`           // 出站规则
	BlockOptions       dbs.JSON `field:"blockOptions"`       // BLOCK动作选项
	PageOptions        dbs.JSON `field:"pageOptions"`        // PAGE动作选项
	CaptchaOptions     dbs.JSON `field:"captchaOptions"`     // 验证码动作选项
	Mode               string   `field:"mode"`               // 模式
	UseLocalFirewall   uint8    `field:"useLocalFirewall"`   // 是否自动使用本地防火墙
	SynFlood           dbs.JSON `field:"synFlood"`           // SynFlood防御设置
	Log                dbs.JSON `field:"log"`                // 日志配置
	MaxRequestBodySize uint32   `field:"maxRequestBodySize"` // 可以检查的最大请求内容尺寸
	DenyCountryHTML    string   `field:"denyCountryHTML"`    // 区域封禁提示
	DenyProvinceHTML   string   `field:"denyProvinceHTML"`   // 省份封禁提示
}

type HTTPFirewallPolicyOperator struct {
	Id                 any // ID
	TemplateId         any // 模版ID
	AdminId            any // 管理员ID
	UserId             any // 用户ID
	ServerId           any // 服务ID
	GroupId            any // 服务分组ID
	State              any // 状态
	CreatedAt          any // 创建时间
	IsOn               any // 是否启用
	Name               any // 名称
	Description        any // 描述
	Inbound            any // 入站规则
	Outbound           any // 出站规则
	BlockOptions       any // BLOCK动作选项
	PageOptions        any // PAGE动作选项
	CaptchaOptions     any // 验证码动作选项
	Mode               any // 模式
	UseLocalFirewall   any // 是否自动使用本地防火墙
	SynFlood           any // SynFlood防御设置
	Log                any // 日志配置
	MaxRequestBodySize any // 可以检查的最大请求内容尺寸
	DenyCountryHTML    any // 区域封禁提示
	DenyProvinceHTML   any // 省份封禁提示
}

func NewHTTPFirewallPolicyOperator() *HTTPFirewallPolicyOperator {
	return &HTTPFirewallPolicyOperator{}
}
