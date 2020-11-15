package events

// 节点更新事件
// TODO 改成事件
var NodeDNSChanges = make(chan int64, 128)
