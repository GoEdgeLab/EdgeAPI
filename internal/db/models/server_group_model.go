package models

// 服务分组
type ServerGroup struct {
	Id        uint32 `field:"id"`        // ID
	AdminId   uint32 `field:"adminId"`   // 管理员ID
	UserId    uint32 `field:"userId"`    // 用户ID
	Name      string `field:"name"`      // 名称
	Order     uint32 `field:"order"`     // 排序
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	State     uint8  `field:"state"`     // 状态
}

type ServerGroupOperator struct {
	Id        interface{} // ID
	AdminId   interface{} // 管理员ID
	UserId    interface{} // 用户ID
	Name      interface{} // 名称
	Order     interface{} // 排序
	CreatedAt interface{} // 创建时间
	State     interface{} // 状态
}

func NewServerGroupOperator() *ServerGroupOperator {
	return &ServerGroupOperator{}
}
