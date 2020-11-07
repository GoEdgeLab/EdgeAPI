package models

// IP
type IPItem struct {
	Id        uint64 `field:"id"`        // ID
	ListId    uint32 `field:"listId"`    // 所属名单ID
	IpFrom    string `field:"ipFrom"`    // 开始IP
	IpTo      string `field:"ipTo"`      // 结束IP
	Version   uint64 `field:"version"`   // 版本
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	UpdatedAt uint64 `field:"updatedAt"` // 修改时间
	Reason    string `field:"reason"`    // 加入说明
	State     uint8  `field:"state"`     // 状态
	ExpiredAt uint64 `field:"expiredAt"` // 过期时间
}

type IPItemOperator struct {
	Id        interface{} // ID
	ListId    interface{} // 所属名单ID
	IpFrom    interface{} // 开始IP
	IpTo      interface{} // 结束IP
	Version   interface{} // 版本
	CreatedAt interface{} // 创建时间
	UpdatedAt interface{} // 修改时间
	Reason    interface{} // 加入说明
	State     interface{} // 状态
	ExpiredAt interface{} // 过期时间
}

func NewIPItemOperator() *IPItemOperator {
	return &IPItemOperator{}
}
