package models

// HTTPBrotliPolicy Gzip配置
type HTTPBrotliPolicy struct {
	Id        uint32 `field:"id"`        // ID
	AdminId   uint32 `field:"adminId"`   // 管理员ID
	UserId    uint32 `field:"userId"`    // 用户ID
	IsOn      uint8  `field:"isOn"`      // 是否启用
	Level     uint32 `field:"level"`     // 压缩级别
	MinLength string `field:"minLength"` // 可压缩最小值
	MaxLength string `field:"maxLength"` // 可压缩最大值
	State     uint8  `field:"state"`     // 状态
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	Conds     string `field:"conds"`     // 条件
}

type HTTPBrotliPolicyOperator struct {
	Id        interface{} // ID
	AdminId   interface{} // 管理员ID
	UserId    interface{} // 用户ID
	IsOn      interface{} // 是否启用
	Level     interface{} // 压缩级别
	MinLength interface{} // 可压缩最小值
	MaxLength interface{} // 可压缩最大值
	State     interface{} // 状态
	CreatedAt interface{} // 创建时间
	Conds     interface{} // 条件
}

func NewHTTPBrotliPolicyOperator() *HTTPBrotliPolicyOperator {
	return &HTTPBrotliPolicyOperator{}
}
