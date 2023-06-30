package models

import "github.com/iwind/TeaGo/dbs"

const (
	LogField_Id              dbs.FieldName = "id"              // ID
	LogField_Level           dbs.FieldName = "level"           // 级别
	LogField_Description     dbs.FieldName = "description"     // 描述
	LogField_CreatedAt       dbs.FieldName = "createdAt"       // 创建时间
	LogField_Action          dbs.FieldName = "action"          // 动作
	LogField_UserId          dbs.FieldName = "userId"          // 用户ID
	LogField_AdminId         dbs.FieldName = "adminId"         // 管理员ID
	LogField_ProviderId      dbs.FieldName = "providerId"      // 供应商ID
	LogField_Ip              dbs.FieldName = "ip"              // IP地址
	LogField_Type            dbs.FieldName = "type"            // 类型：admin, user
	LogField_Day             dbs.FieldName = "day"             // 日期
	LogField_BillId          dbs.FieldName = "billId"          // 账单ID
	LogField_LangMessageCode dbs.FieldName = "langMessageCode" // 多语言消息代号
	LogField_LangMessageArgs dbs.FieldName = "langMessageArgs" // 多语言参数
	LogField_Params          dbs.FieldName = "params"          // 关联对象参数
)

// Log 操作日志
type Log struct {
	Id              uint32   `field:"id"`              // ID
	Level           string   `field:"level"`           // 级别
	Description     string   `field:"description"`     // 描述
	CreatedAt       uint64   `field:"createdAt"`       // 创建时间
	Action          string   `field:"action"`          // 动作
	UserId          uint32   `field:"userId"`          // 用户ID
	AdminId         uint32   `field:"adminId"`         // 管理员ID
	ProviderId      uint32   `field:"providerId"`      // 供应商ID
	Ip              string   `field:"ip"`              // IP地址
	Type            string   `field:"type"`            // 类型：admin, user
	Day             string   `field:"day"`             // 日期
	BillId          uint32   `field:"billId"`          // 账单ID
	LangMessageCode string   `field:"langMessageCode"` // 多语言消息代号
	LangMessageArgs dbs.JSON `field:"langMessageArgs"` // 多语言参数
	Params          dbs.JSON `field:"params"`          // 关联对象参数
}

type LogOperator struct {
	Id              any // ID
	Level           any // 级别
	Description     any // 描述
	CreatedAt       any // 创建时间
	Action          any // 动作
	UserId          any // 用户ID
	AdminId         any // 管理员ID
	ProviderId      any // 供应商ID
	Ip              any // IP地址
	Type            any // 类型：admin, user
	Day             any // 日期
	BillId          any // 账单ID
	LangMessageCode any // 多语言消息代号
	LangMessageArgs any // 多语言参数
	Params          any // 关联对象参数
}

func NewLogOperator() *LogOperator {
	return &LogOperator{}
}
