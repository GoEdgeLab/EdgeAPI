package models

import "github.com/iwind/TeaGo/dbs"

// HTTPCacheTaskKey 缓存任务Key
type HTTPCacheTaskKey struct {
	Id        uint64   `field:"id"`        // ID
	TaskId    uint64   `field:"taskId"`    // 任务ID
	Key       string   `field:"key"`       // Key
	KeyType   string   `field:"keyType"`   // Key类型：key|prefix
	Type      string   `field:"type"`      // 操作类型
	ClusterId uint32   `field:"clusterId"` // 集群ID
	Nodes     dbs.JSON `field:"nodes"`     // 节点
	Errors    dbs.JSON `field:"errors"`    // 错误信息
	IsDone    bool     `field:"isDone"`    // 是否已完成
}

type HTTPCacheTaskKeyOperator struct {
	Id        interface{} // ID
	TaskId    interface{} // 任务ID
	Key       interface{} // Key
	KeyType   interface{} // Key类型：key|prefix
	Type      interface{} // 操作类型
	ClusterId interface{} // 集群ID
	Nodes     interface{} // 节点
	Errors    interface{} // 错误信息
	IsDone    interface{} // 是否已完成
}

func NewHTTPCacheTaskKeyOperator() *HTTPCacheTaskKeyOperator {
	return &HTTPCacheTaskKeyOperator{}
}
