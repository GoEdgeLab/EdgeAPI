package models

// ServerGroup 服务分组
type ServerGroup struct {
	Id               uint32 `field:"id"`               // ID
	AdminId          uint32 `field:"adminId"`          // 管理员ID
	UserId           uint32 `field:"userId"`           // 用户ID
	IsOn             uint8  `field:"isOn"`             // 是否启用
	Name             string `field:"name"`             // 名称
	Order            uint32 `field:"order"`            // 排序
	CreatedAt        uint64 `field:"createdAt"`        // 创建时间
	State            uint8  `field:"state"`            // 状态
	HttpReverseProxy string `field:"httpReverseProxy"` // 反向代理设置
	TcpReverseProxy  string `field:"tcpReverseProxy"`  // TCP反向代理
	UdpReverseProxy  string `field:"udpReverseProxy"`  // UDP反向代理
	WebId            uint32 `field:"webId"`            // Web配置ID
}

type ServerGroupOperator struct {
	Id               interface{} // ID
	AdminId          interface{} // 管理员ID
	UserId           interface{} // 用户ID
	IsOn             interface{} // 是否启用
	Name             interface{} // 名称
	Order            interface{} // 排序
	CreatedAt        interface{} // 创建时间
	State            interface{} // 状态
	HttpReverseProxy interface{} // 反向代理设置
	TcpReverseProxy  interface{} // TCP反向代理
	UdpReverseProxy  interface{} // UDP反向代理
	WebId            interface{} // Web配置ID
}

func NewServerGroupOperator() *ServerGroupOperator {
	return &ServerGroupOperator{}
}
