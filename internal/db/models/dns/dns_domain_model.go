package dns

import "github.com/iwind/TeaGo/dbs"

// DNSDomain 管理的域名
type DNSDomain struct {
	Id            uint32   `field:"id"`            // ID
	AdminId       uint32   `field:"adminId"`       // 管理员ID
	UserId        uint32   `field:"userId"`        // 用户ID
	ProviderId    uint32   `field:"providerId"`    // 服务商ID
	IsOn          bool     `field:"isOn"`          // 是否可用
	Name          string   `field:"name"`          // 域名
	CreatedAt     uint64   `field:"createdAt"`     // 创建时间
	DataUpdatedAt uint64   `field:"dataUpdatedAt"` // 数据更新时间
	DataError     string   `field:"dataError"`     // 数据更新错误
	Data          string   `field:"data"`          // 原始数据信息
	Records       dbs.JSON `field:"records"`       // 所有解析记录
	Routes        dbs.JSON `field:"routes"`        // 线路数据
	IsUp          bool     `field:"isUp"`          // 是否在线
	State         uint8    `field:"state"`         // 状态
	IsDeleted     bool     `field:"isDeleted"`     // 是否已删除
}

type DNSDomainOperator struct {
	Id            interface{} // ID
	AdminId       interface{} // 管理员ID
	UserId        interface{} // 用户ID
	ProviderId    interface{} // 服务商ID
	IsOn          interface{} // 是否可用
	Name          interface{} // 域名
	CreatedAt     interface{} // 创建时间
	DataUpdatedAt interface{} // 数据更新时间
	DataError     interface{} // 数据更新错误
	Data          interface{} // 原始数据信息
	Records       interface{} // 所有解析记录
	Routes        interface{} // 线路数据
	IsUp          interface{} // 是否在线
	State         interface{} // 状态
	IsDeleted     interface{} // 是否已删除
}

func NewDNSDomainOperator() *DNSDomainOperator {
	return &DNSDomainOperator{}
}
