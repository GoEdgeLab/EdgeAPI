package models

// LatestItem 最近的条目统计
type LatestItem struct {
	Id        uint64 `field:"id"`        // ID
	ItemType  string `field:"itemType"`  // Item类型
	ItemId    uint64 `field:"itemId"`    // itemID
	Count     uint64 `field:"count"`     // 数量
	UpdatedAt uint64 `field:"updatedAt"` // 更新时间
}

type LatestItemOperator struct {
	Id        interface{} // ID
	ItemType  interface{} // Item类型
	ItemId    interface{} // itemID
	Count     interface{} // 数量
	UpdatedAt interface{} // 更新时间
}

func NewLatestItemOperator() *LatestItemOperator {
	return &LatestItemOperator{}
}
