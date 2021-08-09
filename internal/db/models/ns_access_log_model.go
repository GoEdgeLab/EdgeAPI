package models

// NSAccessLog 域名服务访问日志
type NSAccessLog struct {
	Id         uint64 `field:"id"`         // ID
	NodeId     uint32 `field:"nodeId"`     // 节点ID
	DomainId   uint32 `field:"domainId"`   // 域名ID
	RecordId   uint32 `field:"recordId"`   // 记录ID
	Content    string `field:"content"`    // 访问数据
	RequestId  string `field:"requestId"`  // 请求ID
	CreatedAt  uint64 `field:"createdAt"`  // 创建时间
	RemoteAddr string `field:"remoteAddr"` // IP
}

type NSAccessLogOperator struct {
	Id         interface{} // ID
	NodeId     interface{} // 节点ID
	DomainId   interface{} // 域名ID
	RecordId   interface{} // 记录ID
	Content    interface{} // 访问数据
	RequestId  interface{} // 请求ID
	CreatedAt  interface{} // 创建时间
	RemoteAddr interface{} // IP
}

func NewNSAccessLogOperator() *NSAccessLogOperator {
	return &NSAccessLogOperator{}
}
