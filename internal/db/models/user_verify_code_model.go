package models

// UserVerifyCode 重置密码之验证码
type UserVerifyCode struct {
	Id         uint64 `field:"id"`         // ID
	Email      string `field:"email"`      // 邮箱地址
	Mobile     string `field:"mobile"`     // 手机号
	Code       string `field:"code"`       // 验证码
	Type       string `field:"type"`       // 类型
	IsSent     bool   `field:"isSent"`     // 是否已发送
	IsVerified bool   `field:"isVerified"` // 是否已激活
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	ExpiresAt  uint64 `field:"expiresAt"`  // 过期时间
	Day        string `field:"day"`        // YYYYMMDD
}

type UserVerifyCodeOperator struct {
	Id         any // ID
	Email      any // 邮箱地址
	Mobile     any // 手机号
	Code       any // 验证码
	Type       any // 类型
	IsSent     any // 是否已发送
	IsVerified any // 是否已激活
	CreatedAt  any // 创建时间
	ExpiresAt  any // 过期时间
	Day        any // YYYYMMDD
}

func NewUserVerifyCodeOperator() *UserVerifyCodeOperator {
	return &UserVerifyCodeOperator{}
}
