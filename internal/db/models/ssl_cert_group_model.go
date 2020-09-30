package models

//
type SSLCertGroup struct {
	Id        uint32 `field:"id"`        // ID
	AdminId   uint32 `field:"adminId"`   // 管理员ID
	UserId    uint32 `field:"userId"`    // 用户ID
	Name      string `field:"name"`      // 分组名
	Order     uint32 `field:"order"`     // 分组排序
	State     uint8  `field:"state"`     // 状态
	CreatedAt uint64 `field:"createdAt"` // 创建时间
}

type SSLCertGroupOperator struct {
	Id        interface{} // ID
	AdminId   interface{} // 管理员ID
	UserId    interface{} // 用户ID
	Name      interface{} // 分组名
	Order     interface{} // 分组排序
	State     interface{} // 状态
	CreatedAt interface{} // 创建时间
}

func NewSSLCertGroupOperator() *SSLCertGroupOperator {
	return &SSLCertGroupOperator{}
}
