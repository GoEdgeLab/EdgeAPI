package models

// 数据库节点
type DBNode struct {
	Id          uint32 `field:"id"`          // ID
	IsOn        uint8  `field:"isOn"`        // 是否启用
	Role        string `field:"role"`        // 数据库角色
	Name        string `field:"name"`        // 名称
	Description string `field:"description"` // 描述
	Host        string `field:"host"`        // 主机
	Port        uint32 `field:"port"`        // 端口
	Username    string `field:"username"`    // 用户名
	Password    string `field:"password"`    // 密码
	Charset     string `field:"charset"`     // 通讯字符集
	ConnTimeout uint32 `field:"connTimeout"` // 连接超时时间（秒）
	State       uint8  `field:"state"`       // 状态
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	Weight      uint32 `field:"weight"`      // 权重
	Order       uint32 `field:"order"`       // 排序
	AdminId     uint32 `field:"adminId"`     // 管理员ID
}

type DBNodeOperator struct {
	Id          interface{} // ID
	IsOn        interface{} // 是否启用
	Role        interface{} // 数据库角色
	Name        interface{} // 名称
	Description interface{} // 描述
	Host        interface{} // 主机
	Port        interface{} // 端口
	Username    interface{} // 用户名
	Password    interface{} // 密码
	Charset     interface{} // 通讯字符集
	ConnTimeout interface{} // 连接超时时间（秒）
	State       interface{} // 状态
	CreatedAt   interface{} // 创建时间
	Weight      interface{} // 权重
	Order       interface{} // 排序
	AdminId     interface{} // 管理员ID
}

func NewDBNodeOperator() *DBNodeOperator {
	return &DBNodeOperator{}
}
