package models

// APIAccessToken API访问令牌
type APIAccessToken struct {
	Id        uint64 `field:"id"`        // ID
	UserId    uint32 `field:"userId"`    // 用户ID
	AdminId   uint32 `field:"adminId"`   // 管理员ID
	Token     string `field:"token"`     // 令牌
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	ExpiredAt uint64 `field:"expiredAt"` // 过期时间
}

type APIAccessTokenOperator struct {
	Id        interface{} // ID
	UserId    interface{} // 用户ID
	AdminId   interface{} // 管理员ID
	Token     interface{} // 令牌
	CreatedAt interface{} // 创建时间
	ExpiredAt interface{} // 过期时间
}

func NewAPIAccessTokenOperator() *APIAccessTokenOperator {
	return &APIAccessTokenOperator{}
}
