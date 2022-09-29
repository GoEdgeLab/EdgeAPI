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
	ClientType  string   `field:"clientType"`  // 客户端类型
	QrcodeTitle string   `field:"qrcodeTitle"` // 二维码标题
	Order       uint32   `field:"order"`       // 排序
	State       uint8    `field:"state"`       // 状态
}

type OrderMethodOperator struct {
	Id          any // ID
	Name        any // 名称
	IsOn        any // 是否启用
	Description any // 描述
	ParentCode  any // 内置的父级代号
	Code        any // 代号
	Url         any // URL
	Secret      any // 密钥
	Params      any // 参数
	ClientType  any // 客户端类型
	QrcodeTitle any // 二维码标题
	Order       any // 排序
	State       any // 状态
}

func NewOrderMethodOperator() *OrderMethodOperator {
	return &OrderMethodOperator{}
}
