package accounts

import "github.com/iwind/TeaGo/dbs"

// OrderMethod 订单支付方式
type OrderMethod struct {
	Id          uint32   `field:"id"`          // ID
	Name        string   `field:"name"`        // 名称
	IsOn        bool     `field:"isOn"`        // 是否启用
	Description string   `field:"description"` // 描述
	ParentCode  string   `field:"parentCode"`  // 内置的父级代号
	Code        string   `field:"code"`        // 代号
	Url         string   `field:"url"`         // URL
	Secret      string   `field:"secret"`      // 密钥
	Params      dbs.JSON `field:"params"`      // 参数
	Order       uint32   `field:"order"`       // 排序
	State       uint8    `field:"state"`       // 状态
}

type OrderMethodOperator struct {
	Id          interface{} // ID
	Name        interface{} // 名称
	IsOn        interface{} // 是否启用
	Description interface{} // 描述
	ParentCode  interface{} // 内置的父级代号
	Code        interface{} // 代号
	Url         interface{} // URL
	Secret      interface{} // 密钥
	Params      interface{} // 参数
	Order       interface{} // 排序
	State       interface{} // 状态
}

func NewOrderMethodOperator() *OrderMethodOperator {
	return &OrderMethodOperator{}
}
