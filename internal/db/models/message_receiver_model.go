package models

import "github.com/iwind/TeaGo/dbs"

// MessageReceiver 消息通知接收人
type MessageReceiver struct {
	Id               uint32   `field:"id"`               // ID
	Role             string   `field:"role"`             // 节点角色
	ClusterId        uint32   `field:"clusterId"`        // 集群ID
	NodeId           uint32   `field:"nodeId"`           // 节点ID
	ServerId         uint32   `field:"serverId"`         // 服务ID
	Type             string   `field:"type"`             // 类型
	Params           dbs.JSON `field:"params"`           // 参数
	RecipientId      uint32   `field:"recipientId"`      // 接收人ID
	RecipientGroupId uint32   `field:"recipientGroupId"` // 接收人分组ID
	State            uint8    `field:"state"`            // 状态
}

type MessageReceiverOperator struct {
	Id               interface{} // ID
	Role             interface{} // 节点角色
	ClusterId        interface{} // 集群ID
	NodeId           interface{} // 节点ID
	ServerId         interface{} // 服务ID
	Type             interface{} // 类型
	Params           interface{} // 参数
	RecipientId      interface{} // 接收人ID
	RecipientGroupId interface{} // 接收人分组ID
	State            interface{} // 状态
}

func NewMessageReceiverOperator() *MessageReceiverOperator {
	return &MessageReceiverOperator{}
}
