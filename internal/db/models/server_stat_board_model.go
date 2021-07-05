package models

// ServerStatBoard 服务统计看板
type ServerStatBoard struct {
	Id        uint64 `field:"id"`        // ID
	Name      string `field:"name"`      // 名称
	ClusterId uint32 `field:"clusterId"` // 集群ID
	IsOn      uint8  `field:"isOn"`      // 是否启用
	Order     uint32 `field:"order"`     // 排序
	State     uint8  `field:"state"`     // 状态
}

type ServerStatBoardOperator struct {
	Id        interface{} // ID
	Name      interface{} // 名称
	ClusterId interface{} // 集群ID
	IsOn      interface{} // 是否启用
	Order     interface{} // 排序
	State     interface{} // 状态
}

func NewServerStatBoardOperator() *ServerStatBoardOperator {
	return &ServerStatBoardOperator{}
}
