package models

import "github.com/iwind/TeaGo/dbs"

const (
	UserScriptField_Id             dbs.FieldName = "id"             // ID
	UserScriptField_UserId         dbs.FieldName = "userId"         // 用户ID
	UserScriptField_AdminId        dbs.FieldName = "adminId"        // 操作管理员
	UserScriptField_Code           dbs.FieldName = "code"           // 代码
	UserScriptField_CodeMD5        dbs.FieldName = "codeMD5"        // 代码MD5
	UserScriptField_CreatedAt      dbs.FieldName = "createdAt"      // 创建时间
	UserScriptField_IsRejected     dbs.FieldName = "isRejected"     // 是否已驳回
	UserScriptField_RejectedAt     dbs.FieldName = "rejectedAt"     // 驳回时间
	UserScriptField_RejectedReason dbs.FieldName = "rejectedReason" // 驳回原因
	UserScriptField_IsPassed       dbs.FieldName = "isPassed"       // 是否通过审核
	UserScriptField_PassedAt       dbs.FieldName = "passedAt"       // 通过时间
	UserScriptField_State          dbs.FieldName = "state"          // 状态
	UserScriptField_WebIds         dbs.FieldName = "webIds"         // WebId列表
)

// UserScript 用户脚本审核
type UserScript struct {
	Id             uint64   `field:"id"`             // ID
	UserId         uint64   `field:"userId"`         // 用户ID
	AdminId        uint64   `field:"adminId"`        // 操作管理员
	Code           string   `field:"code"`           // 代码
	CodeMD5        string   `field:"codeMD5"`        // 代码MD5
	CreatedAt      uint64   `field:"createdAt"`      // 创建时间
	IsRejected     bool     `field:"isRejected"`     // 是否已驳回
	RejectedAt     uint64   `field:"rejectedAt"`     // 驳回时间
	RejectedReason string   `field:"rejectedReason"` // 驳回原因
	IsPassed       bool     `field:"isPassed"`       // 是否通过审核
	PassedAt       uint64   `field:"passedAt"`       // 通过时间
	State          uint8    `field:"state"`          // 状态
	WebIds         dbs.JSON `field:"webIds"`         // WebId列表
}

type UserScriptOperator struct {
	Id             any // ID
	UserId         any // 用户ID
	AdminId        any // 操作管理员
	Code           any // 代码
	CodeMD5        any // 代码MD5
	CreatedAt      any // 创建时间
	IsRejected     any // 是否已驳回
	RejectedAt     any // 驳回时间
	RejectedReason any // 驳回原因
	IsPassed       any // 是否通过审核
	PassedAt       any // 通过时间
	State          any // 状态
	WebIds         any // WebId列表
}

func NewUserScriptOperator() *UserScriptOperator {
	return &UserScriptOperator{}
}
