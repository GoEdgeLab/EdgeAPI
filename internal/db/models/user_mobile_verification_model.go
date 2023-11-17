package models

import "github.com/iwind/TeaGo/dbs"

const (
	UserMobileVerificationField_Id dbs.FieldName = "id"         // ID
	UserMobileVerificationField_Mobile     dbs.FieldName = "mobile"     // 手机号码
	UserMobileVerificationField_UserId     dbs.FieldName = "userId"     // 用户ID
	UserMobileVerificationField_Code       dbs.FieldName = "code"       // 激活码
	UserMobileVerificationField_CreatedAt  dbs.FieldName = "createdAt"  // 创建时间
	UserMobileVerificationField_IsSent     dbs.FieldName = "isSent"     // 是否已发送
	UserMobileVerificationField_IsVerified dbs.FieldName = "isVerified" // 是否已激活
	UserMobileVerificationField_Day        dbs.FieldName = "day"        // YYYYMMDD
)

// UserMobileVerification 邮箱激活邮件队列
type UserMobileVerification struct {
	Id         uint64 `field:"id"`         // ID
	Mobile     string `field:"mobile"`     // 手机号码
	UserId     uint64 `field:"userId"`     // 用户ID
	Code       string `field:"code"`       // 激活码
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	IsSent     bool   `field:"isSent"`     // 是否已发送
	IsVerified bool   `field:"isVerified"` // 是否已激活
	Day        string `field:"day"`        // YYYYMMDD
}

type UserMobileVerificationOperator struct {
	Id         any // ID
	Mobile     any // 手机号码
	UserId     any // 用户ID
	Code       any // 激活码
	CreatedAt  any // 创建时间
	IsSent     any // 是否已发送
	IsVerified any // 是否已激活
	Day        any // YYYYMMDD
}

func NewUserMobileVerificationOperator() *UserMobileVerificationOperator {
	return &UserMobileVerificationOperator{}
}
