package nameservers

// NSDomainGroup 域名分组
type NSDomainGroup struct {
	Id     uint64 `field:"id"`     // ID
	UserId uint64 `field:"userId"` // 用户ID
	Name   string `field:"name"`   // 分组名称
	IsOn   bool   `field:"isOn"`   // 是否启用
	Order  uint32 `field:"order"`  // 排序
	State  uint8  `field:"state"`  // 状态
}

type NSDomainGroupOperator struct {
	Id     interface{} // ID
	UserId interface{} // 用户ID
	Name   interface{} // 分组名称
	IsOn   interface{} // 是否启用
	Order  interface{} // 排序
	State  interface{} // 状态
}

func NewNSDomainGroupOperator() *NSDomainGroupOperator {
	return &NSDomainGroupOperator{}
}
