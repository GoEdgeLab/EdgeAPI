package clients

// ClientAgentIP Agent IP
type ClientAgentIP struct {
	Id      uint64 `field:"id"`      // ID
	AgentId uint32 `field:"agentId"` // Agent ID
	IP      string `field:"ip"`      // IP地址
	Ptr     string `field:"ptr"`     // PTR值
}

type ClientAgentIPOperator struct {
	Id      any // ID
	AgentId any // Agent ID
	IP      any // IP地址
	Ptr     any // PTR值
}

func NewClientAgentIPOperator() *ClientAgentIPOperator {
	return &ClientAgentIPOperator{}
}
