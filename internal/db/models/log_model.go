package models

import "github.com/iwind/TeaGo/dbs"

const (
	LogFieldId              dbs.FieldName = "id"              // ID
	LogFieldLevel           dbs.FieldName = "level"           // 级别
	LogFieldDescription     dbs.FieldName = "description"     // 描述
	LogFieldCreatedAt       dbs.FieldName = "createdAt"       // 创建时间
	LogFieldAction          dbs.FieldName = "action"          // 动作
	LogFieldUserId          dbs.FieldName = "userId"          // 用户ID
	LogFieldAdminId         dbs.FieldName = "adminId"         // 管理员ID
	LogFieldProviderId      dbs.FieldName = "providerId"      // 供应商ID
	LogFieldIp              dbs.FieldName = "ip"              // IP地址
	LogFieldType            dbs.FieldName = "type"            // 类型：admin, user
	LogFieldDay             dbs.FieldName = "day"             // 日期
	LogFieldBillId          dbs.FieldName = "billId"          // 账单ID
	LogFieldLangMessageCode dbs.FieldName = "langMessageCode" // 多语言消息代号
	LogFieldLangMesageArgs  dbs.FieldName = "langMesageArgs"  // 多语言参数
	LogFieldParams          dbs.FieldName = "params"          // 关联对象参数
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
	LangMesageArgs  dbs.JSON `field:"langMesageArgs"`  // 多语言参数
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
	LangMesageArgs  any // 多语言参数
	Params          any // 关联对象参数
}

func NewLogOperator() *LogOperator {
	return &LogOperator{}
}
