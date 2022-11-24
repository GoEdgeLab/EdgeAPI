package nameservers

// NSRouteCategory 线路分类
type NSRouteCategory struct {
	Id      uint64 `field:"id"`      // ID
	IsOn    bool   `field:"isOn"`    // 是否启用
	Name    string `field:"name"`    // 分类名
	AdminId uint64 `field:"adminId"` // 管理员ID
	UserId  uint64 `field:"userId"`  // 用户ID
	Order   uint32 `field:"order"`   // 排序
	State   uint8  `field:"state"`   // 状态
}

type NSRouteCategoryOperator struct {
	Id      any // ID
	IsOn    any // 是否启用
	Name    any // 分类名
	AdminId any // 管理员ID
	UserId  any // 用户ID
	Order   any // 排序
	State   any // 状态
}

func NewNSRouteCategoryOperator() *NSRouteCategoryOperator {
	return &NSRouteCategoryOperator{}
}
