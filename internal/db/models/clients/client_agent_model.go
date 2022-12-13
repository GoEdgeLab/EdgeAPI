package clients

// ClientAgent Agent库
type ClientAgent struct {
	Id          uint32 `field:"id"`          // ID
	Name        string `field:"name"`        // 名称
	Code        string `field:"code"`        // 代号
	Description string `field:"description"` // 介绍
	Order       uint32 `field:"order"`       // 排序
	CountIPs    uint32 `field:"countIPs"`    // IP数量
}

type ClientAgentOperator struct {
	Id          any // ID
	Name        any // 名称
	Code        any // 代号
	Description any // 介绍
	Order       any // 排序
	CountIPs    any // IP数量
}

func NewClientAgentOperator() *ClientAgentOperator {
	return &ClientAgentOperator{}
}
