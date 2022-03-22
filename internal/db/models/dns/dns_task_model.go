package dns

// DNS更新任务
type DNSTask struct {
	Id        uint64 `field:"id"`        // ID
	ClusterId uint32 `field:"clusterId"` // 集群ID
	ServerId  uint32 `field:"serverId"`  // 服务ID
	NodeId    uint32 `field:"nodeId"`    // 节点ID
	DomainId  uint32 `field:"domainId"`  // 域名ID
	Type      string `field:"type"`      // 任务类型
	UpdatedAt uint64 `field:"updatedAt"` // 更新时间
	IsDone    bool   `field:"isDone"`    // 是否已完成
	IsOk      bool   `field:"isOk"`      // 是否成功
	Error     string `field:"error"`     // 错误信息
}

type DNSTaskOperator struct {
	Id        interface{} // ID
	ClusterId interface{} // 集群ID
	ServerId  interface{} // 服务ID
	NodeId    interface{} // 节点ID
	DomainId  interface{} // 域名ID
	Type      interface{} // 任务类型
	UpdatedAt interface{} // 更新时间
	IsDone    interface{} // 是否已完成
	IsOk      interface{} // 是否成功
	Error     interface{} // 错误信息
}

func NewDNSTaskOperator() *DNSTaskOperator {
	return &DNSTaskOperator{}
}
