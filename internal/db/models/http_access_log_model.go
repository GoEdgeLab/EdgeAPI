package models

//
type HTTPAccessLog struct {
	Id        uint64 `field:"id"`        // ID
	ServerId  uint32 `field:"serverId"`  // 服务ID
	NodeId    uint32 `field:"nodeId"`    // 节点ID
	Status    uint32 `field:"status"`    // 状态码
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	Content   string `field:"content"`   // 日志内容
	RequestId string `field:"requestId"` // 请求ID
}

type HTTPAccessLogOperator struct {
	Id        interface{} // ID
	ServerId  interface{} // 服务ID
	NodeId    interface{} // 节点ID
	Status    interface{} // 状态码
	CreatedAt interface{} // 创建时间
	Content   interface{} // 日志内容
	RequestId interface{} // 请求ID
}

func NewHTTPAccessLogOperator() *HTTPAccessLogOperator {
	return &HTTPAccessLogOperator{}
}
