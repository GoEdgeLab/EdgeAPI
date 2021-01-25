package acme

// ACME认证
type ACMEAuthentication struct {
	Id        uint64 `field:"id"`        // ID
	TaskId    uint64 `field:"taskId"`    // 任务ID
	Domain    string `field:"domain"`    // 域名
	Token     string `field:"token"`     // 令牌
	Key       string `field:"key"`       // 密钥
	CreatedAt uint64 `field:"createdAt"` // 创建时间
}

type ACMEAuthenticationOperator struct {
	Id        interface{} // ID
	TaskId    interface{} // 任务ID
	Domain    interface{} // 域名
	Token     interface{} // 令牌
	Key       interface{} // 密钥
	CreatedAt interface{} // 创建时间
}

func NewACMEAuthenticationOperator() *ACMEAuthenticationOperator {
	return &ACMEAuthenticationOperator{}
}
