package models

// Message 消息通知
type Message struct {
	Id        uint64 `field:"id"`        // ID
	AdminId   uint32 `field:"adminId"`   // 管理员ID
	UserId    uint32 `field:"userId"`    // 用户ID
	ClusterId uint32 `field:"clusterId"` // 集群ID
	NodeId    uint32 `field:"nodeId"`    // 节点ID
	Level     string `field:"level"`     // 级别
	Subject   string `field:"subject"`   // 标题
	Body      string `field:"body"`      // 内容
	Type      string `field:"type"`      // 消息类型
	Params    string `field:"params"`    // 额外的参数
	IsRead    uint8  `field:"isRead"`    // 是否已读
	State     uint8  `field:"state"`     // 状态
	CreatedAt uint64 `field:"createdAt"` // 创建时间
	Day       string `field:"day"`       // 日期YYYYMMDD
	Hash      string `field:"hash"`      // 消息内容的Hash
}

type MessageOperator struct {
	Id        interface{} // ID
	AdminId   interface{} // 管理员ID
	UserId    interface{} // 用户ID
	ClusterId interface{} // 集群ID
	NodeId    interface{} // 节点ID
	Level     interface{} // 级别
	Subject   interface{} // 标题
	Body      interface{} // 内容
	Type      interface{} // 消息类型
	Params    interface{} // 额外的参数
	IsRead    interface{} // 是否已读
	State     interface{} // 状态
	CreatedAt interface{} // 创建时间
	Day       interface{} // 日期YYYYMMDD
	Hash      interface{} // 消息内容的Hash
}

func NewMessageOperator() *MessageOperator {
	return &MessageOperator{}
}
