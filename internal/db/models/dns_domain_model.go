package models

// 管理的域名
type DNSDomain struct {
	Id            uint32 `field:"id"`            // ID
	AdminId       uint32 `field:"adminId"`       // 管理员ID
	ProviderId    uint32 `field:"providerId"`    // 服务商ID
	IsOn          uint8  `field:"isOn"`          // 是否可用
	Name          string `field:"name"`          // 域名
	CreatedAt     uint64 `field:"createdAt"`     // 创建时间
	DataUpdatedAt uint64 `field:"dataUpdatedAt"` // 数据更新时间
	DataError     string `field:"dataError"`     // 数据更新错误
	Data          string `field:"data"`          // 原始数据信息
	Records       string `field:"records"`       // 所有解析记录
	Routes        string `field:"routes"`        // 线路数据
	State         uint8  `field:"state"`         // 状态
}

type DNSDomainOperator struct {
	Id            interface{} // ID
	AdminId       interface{} // 管理员ID
	ProviderId    interface{} // 服务商ID
	IsOn          interface{} // 是否可用
	Name          interface{} // 域名
	CreatedAt     interface{} // 创建时间
	DataUpdatedAt interface{} // 数据更新时间
	DataError     interface{} // 数据更新错误
	Data          interface{} // 原始数据信息
	Records       interface{} // 所有解析记录
	Routes        interface{} // 线路数据
	State         interface{} // 状态
}

func NewDNSDomainOperator() *DNSDomainOperator {
	return &DNSDomainOperator{}
}
