package models

import "github.com/iwind/TeaGo/dbs"

// UserNode 用户节点
type UserNode struct {
	Id          uint32   `field:"id"`          // ID
	IsOn        bool     `field:"isOn"`        // 是否启用
	UniqueId    string   `field:"uniqueId"`    // 唯一ID
	Secret      string   `field:"secret"`      // 密钥
	Name        string   `field:"name"`        // 名称
	Description string   `field:"description"` // 描述
	Http        dbs.JSON `field:"http"`        // 监听的HTTP配置
	Https       dbs.JSON `field:"https"`       // 监听的HTTPS配置
	AccessAddrs dbs.JSON `field:"accessAddrs"` // 外部访问地址
	Order       uint32   `field:"order"`       // 排序
	State       uint8    `field:"state"`       // 状态
	CreatedAt   uint64   `field:"createdAt"`   // 创建时间
	AdminId     uint32   `field:"adminId"`     // 管理员ID
	Weight      uint32   `field:"weight"`      // 权重
	Status      dbs.JSON `field:"status"`      // 运行状态
}

type UserNodeOperator struct {
	Id          interface{} // ID
	IsOn        interface{} // 是否启用
	UniqueId    interface{} // 唯一ID
	Secret      interface{} // 密钥
	Name        interface{} // 名称
	Description interface{} // 描述
	Http        interface{} // 监听的HTTP配置
	Https       interface{} // 监听的HTTPS配置
	AccessAddrs interface{} // 外部访问地址
	Order       interface{} // 排序
	State       interface{} // 状态
	CreatedAt   interface{} // 创建时间
	AdminId     interface{} // 管理员ID
	Weight      interface{} // 权重
	Status      interface{} // 运行状态
}

func NewUserNodeOperator() *UserNodeOperator {
	return &UserNodeOperator{}
}
