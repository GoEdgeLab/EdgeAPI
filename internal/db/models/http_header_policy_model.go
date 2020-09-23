package models

//
type HTTPHeaderPolicy struct {
	Id             uint32 `field:"id"`             // ID
	IsOn           uint8  `field:"isOn"`           // 是否启用
	State          uint8  `field:"state"`          // 状态
	AdminId        uint32 `field:"adminId"`        // 管理员ID
	UserId         uint32 `field:"userId"`         // 用户ID
	CreatedAt      uint64 `field:"createdAt"`      // 创建时间
	AddHeaders     string `field:"addHeaders"`     // 添加的Header
	AddTrailers    string `field:"addTrailers"`    // 添加的Trailers
	SetHeaders     string `field:"setHeaders"`     // 设置Header
	ReplaceHeaders string `field:"replaceHeaders"` // 替换Header内容
	Expires        string `field:"expires"`        // Expires单独设置
	DeleteHeaders  string `field:"deleteHeaders"`  // 删除的Headers
}

type HTTPHeaderPolicyOperator struct {
	Id             interface{} // ID
	IsOn           interface{} // 是否启用
	State          interface{} // 状态
	AdminId        interface{} // 管理员ID
	UserId         interface{} // 用户ID
	CreatedAt      interface{} // 创建时间
	AddHeaders     interface{} // 添加的Header
	AddTrailers    interface{} // 添加的Trailers
	SetHeaders     interface{} // 设置Header
	ReplaceHeaders interface{} // 替换Header内容
	Expires        interface{} // Expires单独设置
	DeleteHeaders  interface{} // 删除的Headers
}

func NewHTTPHeaderPolicyOperator() *HTTPHeaderPolicyOperator {
	return &HTTPHeaderPolicyOperator{}
}
