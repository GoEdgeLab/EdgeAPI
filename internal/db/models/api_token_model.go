package models

// API令牌管理
type ApiToken struct {
	Id     uint32 `field:"id"`     // ID
	NodeId string `field:"nodeId"` // 节点ID
	Secret string `field:"secret"` // 节点密钥
	Role   string `field:"role"`   // 节点角色
	State  uint8  `field:"state"`  // 状态
}

type ApiTokenOperator struct {
	Id     interface{} // ID
	NodeId interface{} // 节点ID
	Secret interface{} // 节点密钥
	Role   interface{} // 节点角色
	State  interface{} // 状态
}

func NewApiTokenOperator() *ApiTokenOperator {
	return &ApiTokenOperator{}
}
