package models

import "github.com/iwind/TeaGo/dbs"

const (
	ServerBandwidthStatField_Id                  dbs.FieldName = "id"                  // ID
	ServerBandwidthStatField_UserId              dbs.FieldName = "userId"              // 用户ID
	ServerBandwidthStatField_ServerId            dbs.FieldName = "serverId"            // 服务ID
	ServerBandwidthStatField_RegionId            dbs.FieldName = "regionId"            // 区域ID
	ServerBandwidthStatField_UserPlanId          dbs.FieldName = "userPlanId"          // 用户套餐ID
	ServerBandwidthStatField_Day                 dbs.FieldName = "day"                 // 日期YYYYMMDD
	ServerBandwidthStatField_TimeAt              dbs.FieldName = "timeAt"              // 时间点HHMM
	ServerBandwidthStatField_Bytes               dbs.FieldName = "bytes"               // 带宽字节
	ServerBandwidthStatField_AvgBytes            dbs.FieldName = "avgBytes"            // 平均流量
	ServerBandwidthStatField_CachedBytes         dbs.FieldName = "cachedBytes"         // 缓存的流量
	ServerBandwidthStatField_AttackBytes         dbs.FieldName = "attackBytes"         // 攻击流量
	ServerBandwidthStatField_CountRequests       dbs.FieldName = "countRequests"       // 请求数
	ServerBandwidthStatField_CountCachedRequests dbs.FieldName = "countCachedRequests" // 缓存的请求数
	ServerBandwidthStatField_CountAttackRequests dbs.FieldName = "countAttackRequests" // 攻击请求数
	ServerBandwidthStatField_TotalBytes          dbs.FieldName = "totalBytes"          // 总流量
	ServerBandwidthStatField_CountIPs            dbs.FieldName = "countIPs"            // 独立IP
)

// ServerBandwidthStat 服务峰值带宽统计
type ServerBandwidthStat struct {
	Id                  uint64 `field:"id"`                  // ID
	UserId              uint64 `field:"userId"`              // 用户ID
	ServerId            uint64 `field:"serverId"`            // 服务ID
	RegionId            uint32 `field:"regionId"`            // 区域ID
	UserPlanId          uint64 `field:"userPlanId"`          // 用户套餐ID
	Day                 string `field:"day"`                 // 日期YYYYMMDD
	TimeAt              string `field:"timeAt"`              // 时间点HHMM
	Bytes               uint64 `field:"bytes"`               // 带宽字节
	AvgBytes            uint64 `field:"avgBytes"`            // 平均流量
	CachedBytes         uint64 `field:"cachedBytes"`         // 缓存的流量
	AttackBytes         uint64 `field:"attackBytes"`         // 攻击流量
	CountRequests       uint64 `field:"countRequests"`       // 请求数
	CountCachedRequests uint64 `field:"countCachedRequests"` // 缓存的请求数
	CountAttackRequests uint64 `field:"countAttackRequests"` // 攻击请求数
	TotalBytes          uint64 `field:"totalBytes"`          // 总流量
	CountIPs            uint64 `field:"countIPs"`            // 独立IP
}

type ServerBandwidthStatOperator struct {
	Id                  any // ID
	UserId              any // 用户ID
	ServerId            any // 服务ID
	RegionId            any // 区域ID
	UserPlanId          any // 用户套餐ID
	Day                 any // 日期YYYYMMDD
	TimeAt              any // 时间点HHMM
	Bytes               any // 带宽字节
	AvgBytes            any // 平均流量
	CachedBytes         any // 缓存的流量
	AttackBytes         any // 攻击流量
	CountRequests       any // 请求数
	CountCachedRequests any // 缓存的请求数
	CountAttackRequests any // 攻击请求数
	TotalBytes          any // 总流量
	CountIPs            any // 独立IP
}

func NewServerBandwidthStatOperator() *ServerBandwidthStatOperator {
	return &ServerBandwidthStatOperator{}
}
