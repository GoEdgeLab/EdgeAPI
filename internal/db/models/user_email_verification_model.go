package models

// UserEmailVerification 邮箱激活邮件队列
type UserEmailVerification struct {
	Id         uint64 `field:"id"`         // ID
	Email      string `field:"email"`      // 邮箱
	UserId     uint64 `field:"userId"`     // 用户ID
	Code       string `field:"code"`       // 激活码
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	IsSent     bool   `field:"isSent"`     // 是否已发送
	IsVerified bool   `field:"isVerified"` // 是否已激活
	Day        string `field:"day"`        // YYYYMMDD
}

type UserEmailVerificationOperator struct {
	Id         any // ID
	Email      any // 邮箱
	UserId     any // 用户ID
	Code       any // 激活码
	CreatedAt  any // 创建时间
	IsSent     any // 是否已发送
	IsVerified any // 是否已激活
	Day        any // YYYYMMDD
}

func NewUserEmailVerificationOperator() *UserEmailVerificationOperator {
	return &UserEmailVerificationOperator{}
}
