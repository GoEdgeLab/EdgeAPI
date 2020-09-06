package models

//
type APINode struct {
	Id          uint32 `field:"id"`          // ID
	IsOn        uint8  `field:"isOn"`        // 是否启用
	ClusterId   uint32 `field:"clusterId"`   // 专用集群ID
	UniqueId    string `field:"uniqueId"`    // 唯一ID
	Secret      string `field:"secret"`      // 密钥
	Name        string `field:"name"`        // 名称
	Description string `field:"description"` // 描述
	Host        string `field:"host"`        // 地址
	Port        uint32 `field:"port"`        // 端口
	Order       uint32 `field:"order"`       // 排序
	State       uint8  `field:"state"`       // 状态
}

type APINodeOperator struct {
	Id          interface{} // ID
	IsOn        interface{} // 是否启用
	ClusterId   interface{} // 专用集群ID
	UniqueId    interface{} // 唯一ID
	Secret      interface{} // 密钥
	Name        interface{} // 名称
	Description interface{} // 描述
	Host        interface{} // 地址
	Port        interface{} // 端口
	Order       interface{} // 排序
	State       interface{} // 状态
}

func NewAPINodeOperator() *APINodeOperator {
	return &APINodeOperator{}
}
