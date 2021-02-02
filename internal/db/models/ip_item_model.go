package models

// IP
type IPItem struct {
	Id         uint64 `field:"id"`         // ID
	ListId     uint32 `field:"listId"`     // 所属名单ID
	Type       string `field:"type"`       // 类型
	IpFrom     string `field:"ipFrom"`     // 开始IP
	IpTo       string `field:"ipTo"`       // 结束IP
	IpFromLong uint64 `field:"ipFromLong"` // 开始IP整型
	IpToLong   uint64 `field:"ipToLong"`   // 结束IP整型
	Version    uint64 `field:"version"`    // 版本
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	UpdatedAt  uint64 `field:"updatedAt"`  // 修改时间
	Reason     string `field:"reason"`     // 加入说明
	Action     string `field:"action"`     // 动作代号
	State      uint8  `field:"state"`      // 状态
	ExpiredAt  uint64 `field:"expiredAt"`  // 过期时间
}

type IPItemOperator struct {
	Id         interface{} // ID
	ListId     interface{} // 所属名单ID
	Type       interface{} // 类型
	IpFrom     interface{} // 开始IP
	IpTo       interface{} // 结束IP
	IpFromLong interface{} // 开始IP整型
	IpToLong   interface{} // 结束IP整型
	Version    interface{} // 版本
	CreatedAt  interface{} // 创建时间
	UpdatedAt  interface{} // 修改时间
	Reason     interface{} // 加入说明
	Action     interface{} // 动作代号
	State      interface{} // 状态
	ExpiredAt  interface{} // 过期时间
}

func NewIPItemOperator() *IPItemOperator {
	return &IPItemOperator{}
}
