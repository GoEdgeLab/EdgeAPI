package models

import "github.com/iwind/TeaGo/dbs"

const (
	UserPlanStatField_Id                        dbs.FieldName = "id"                        // ID
	UserPlanStatField_UserPlanId                dbs.FieldName = "userPlanId"                // 用户套餐ID
	UserPlanStatField_Date                      dbs.FieldName = "date"                      // 日期：YYYYMMDD或YYYYMM
	UserPlanStatField_DateType                  dbs.FieldName = "dateType"                  // 日期类型：day|month
	UserPlanStatField_TrafficBytes              dbs.FieldName = "trafficBytes"              // 流量
	UserPlanStatField_CountRequests             dbs.FieldName = "countRequests"             // 总请求数
	UserPlanStatField_CountWebsocketConnections dbs.FieldName = "countWebsocketConnections" // Websocket连接数
	UserPlanStatField_IsProcessed               dbs.FieldName = "isProcessed"               // 是否已处理
)

// UserPlanStat 用户套餐统计
type UserPlanStat struct {
	Id                        uint64 `field:"id"`                        // ID
	UserPlanId                uint64 `field:"userPlanId"`                // 用户套餐ID
	Date                      string `field:"date"`                      // 日期：YYYYMMDD或YYYYMM
	DateType                  string `field:"dateType"`                  // 日期类型：day|month
	TrafficBytes              uint64 `field:"trafficBytes"`              // 流量
	CountRequests             uint64 `field:"countRequests"`             // 总请求数
	CountWebsocketConnections uint64 `field:"countWebsocketConnections"` // Websocket连接数
	IsProcessed               bool   `field:"isProcessed"`               // 是否已处理
}

type UserPlanStatOperator struct {
	Id                        any // ID
	UserPlanId                any // 用户套餐ID
	Date                      any // 日期：YYYYMMDD或YYYYMM
	DateType                  any // 日期类型：day|month
	TrafficBytes              any // 流量
	CountRequests             any // 总请求数
	CountWebsocketConnections any // Websocket连接数
	IsProcessed               any // 是否已处理
}

func NewUserPlanStatOperator() *UserPlanStatOperator {
	return &UserPlanStatOperator{}
}
