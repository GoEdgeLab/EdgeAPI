package events

type Event = string

const (
	EventStart  Event = "start"  // start loading
	EventLoaded Event = "loaded" // first load
	EventQuit   Event = "quit"   // quit node gracefully
	EventReload Event = "reload" // reload config
)

// 节点更新事件
// TODO 改成事件
var NodeDNSChanges = make(chan int64, 128)
