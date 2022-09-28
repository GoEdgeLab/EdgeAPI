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
	Id          any // ID
	IsOn        any // 是否启用
	UniqueId    any // 唯一ID
	Secret      any // 密钥
	Name        any // 名称
	Description any // 描述
	Http        any // 监听的HTTP配置
	Https       any // 监听的HTTPS配置
	AccessAddrs any // 外部访问地址
	Order       any // 排序
	State       any // 状态
	CreatedAt   any // 创建时间
	AdminId     any // 管理员ID
	Weight      any // 权重
	Status      any // 运行状态
}

func NewUserNodeOperator() *UserNodeOperator {
	return &UserNodeOperator{}
}
