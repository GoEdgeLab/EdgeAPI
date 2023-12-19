package models

import "github.com/iwind/TeaGo/dbs"

const (
	PlanField_Id                          dbs.FieldName = "id"                          // ID
	PlanField_IsOn                        dbs.FieldName = "isOn"                        // 是否启用
	PlanField_Name                        dbs.FieldName = "name"                        // 套餐名
	PlanField_ClusterId                   dbs.FieldName = "clusterId"                   // 集群ID
	PlanField_TrafficLimit                dbs.FieldName = "trafficLimit"                // 流量限制
	PlanField_Features                    dbs.FieldName = "features"                    // 允许的功能
	PlanField_TrafficPrice                dbs.FieldName = "trafficPrice"                // 流量价格设定
	PlanField_BandwidthPrice              dbs.FieldName = "bandwidthPrice"              // 带宽价格
	PlanField_MonthlyPrice                dbs.FieldName = "monthlyPrice"                // 月付
	PlanField_SeasonallyPrice             dbs.FieldName = "seasonallyPrice"             // 季付
	PlanField_YearlyPrice                 dbs.FieldName = "yearlyPrice"                 // 年付
	PlanField_PriceType                   dbs.FieldName = "priceType"                   // 价格类型
	PlanField_Order                       dbs.FieldName = "order"                       // 排序
	PlanField_State                       dbs.FieldName = "state"                       // 状态
	PlanField_TotalServers                dbs.FieldName = "totalServers"                // 可以绑定的网站数量
	PlanField_TotalServerNamesPerServer   dbs.FieldName = "totalServerNamesPerServer"   // 每个网站可以绑定的域名数量
	PlanField_TotalServerNames            dbs.FieldName = "totalServerNames"            // 总域名数量
	PlanField_MonthlyRequests             dbs.FieldName = "monthlyRequests"             // 每月访问量额度
	PlanField_DailyRequests               dbs.FieldName = "dailyRequests"               // 每日访问量额度
	PlanField_DailyWebsocketConnections   dbs.FieldName = "dailyWebsocketConnections"   // 每日Websocket连接数
	PlanField_MonthlyWebsocketConnections dbs.FieldName = "monthlyWebsocketConnections" // 每月Websocket连接数
)

// Plan 用户套餐
type Plan struct {
	Id                          uint32   `field:"id"`                          // ID
	IsOn                        bool     `field:"isOn"`                        // 是否启用
	Name                        string   `field:"name"`                        // 套餐名
	ClusterId                   uint32   `field:"clusterId"`                   // 集群ID
	TrafficLimit                dbs.JSON `field:"trafficLimit"`                // 流量限制
	Features                    dbs.JSON `field:"features"`                    // 允许的功能
	TrafficPrice                dbs.JSON `field:"trafficPrice"`                // 流量价格设定
	BandwidthPrice              dbs.JSON `field:"bandwidthPrice"`              // 带宽价格
	MonthlyPrice                float64  `field:"monthlyPrice"`                // 月付
	SeasonallyPrice             float64  `field:"seasonallyPrice"`             // 季付
	YearlyPrice                 float64  `field:"yearlyPrice"`                 // 年付
	PriceType                   string   `field:"priceType"`                   // 价格类型
	Order                       uint32   `field:"order"`                       // 排序
	State                       uint8    `field:"state"`                       // 状态
	TotalServers                uint32   `field:"totalServers"`                // 可以绑定的网站数量
	TotalServerNamesPerServer   uint32   `field:"totalServerNamesPerServer"`   // 每个网站可以绑定的域名数量
	TotalServerNames            uint32   `field:"totalServerNames"`            // 总域名数量
	MonthlyRequests             uint64   `field:"monthlyRequests"`             // 每月访问量额度
	DailyRequests               uint64   `field:"dailyRequests"`               // 每日访问量额度
	DailyWebsocketConnections   uint64   `field:"dailyWebsocketConnections"`   // 每日Websocket连接数
	MonthlyWebsocketConnections uint64   `field:"monthlyWebsocketConnections"` // 每月Websocket连接数
}

type PlanOperator struct {
	Id                          any // ID
	IsOn                        any // 是否启用
	Name                        any // 套餐名
	ClusterId                   any // 集群ID
	TrafficLimit                any // 流量限制
	Features                    any // 允许的功能
	TrafficPrice                any // 流量价格设定
	BandwidthPrice              any // 带宽价格
	MonthlyPrice                any // 月付
	SeasonallyPrice             any // 季付
	YearlyPrice                 any // 年付
	PriceType                   any // 价格类型
	Order                       any // 排序
	State                       any // 状态
	TotalServers                any // 可以绑定的网站数量
	TotalServerNamesPerServer   any // 每个网站可以绑定的域名数量
	TotalServerNames            any // 总域名数量
	MonthlyRequests             any // 每月访问量额度
	DailyRequests               any // 每日访问量额度
	DailyWebsocketConnections   any // 每日Websocket连接数
	MonthlyWebsocketConnections any // 每月Websocket连接数
}

func NewPlanOperator() *PlanOperator {
	return &PlanOperator{}
}
