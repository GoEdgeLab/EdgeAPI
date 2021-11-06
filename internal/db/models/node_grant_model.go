package models

// NodeGrant 节点授权
type NodeGrant struct {
	Id          uint32 `field:"id"`          // ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	Name        string `field:"name"`        // 名称
	Method      string `field:"method"`      // 登录方式
	Username    string `field:"username"`    // 用户名
	Password    string `field:"password"`    // 密码
	Su          uint8  `field:"su"`          // 是否需要su
	PrivateKey  string `field:"privateKey"`  // 私钥
	Passphrase  string `field:"passphrase"`  // 私钥密码
	Description string `field:"description"` // 备注
	NodeId      uint32 `field:"nodeId"`      // 专有节点
	Role        string `field:"role"`        // 角色
	State       uint8  `field:"state"`       // 状态
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
}

type NodeGrantOperator struct {
	Id          interface{} // ID
	AdminId     interface{} // 管理员ID
	Name        interface{} // 名称
	Method      interface{} // 登录方式
	Username    interface{} // 用户名
	Password    interface{} // 密码
	Su          interface{} // 是否需要su
	PrivateKey  interface{} // 私钥
	Passphrase  interface{} // 私钥密码
	Description interface{} // 备注
	NodeId      interface{} // 专有节点
	Role        interface{} // 角色
	State       interface{} // 状态
	CreatedAt   interface{} // 创建时间
}

func NewNodeGrantOperator() *NodeGrantOperator {
	return &NodeGrantOperator{}
}
