package models

// ADNetwork 高防线路
type ADNetwork struct {
	Id          uint32 `field:"id"`          // ID
	IsOn        bool   `field:"isOn"`        // 是否启用
	Name        string `field:"name"`        // 名称
	Description string `field:"description"` // 描述
	Order       uint32 `field:"order"`       // 排序
	State       uint8  `field:"state"`       // 状态
}

type ADNetworkOperator struct {
	Id          any // ID
	IsOn        any // 是否启用
	Name        any // 名称
	Description any // 描述
	Order       any // 排序
	State       any // 状态
}

func NewADNetworkOperator() *ADNetworkOperator {
	return &ADNetworkOperator{}
}
