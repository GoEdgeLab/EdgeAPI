package models

// UserTrafficPackage 用户购买的流量包
type UserTrafficPackage struct {
	Id          uint64 `field:"id"`          // ID
	AdminId     uint32 `field:"adminId"`     // 管理员ID
	UserId      uint64 `field:"userId"`      // 用户ID
	PackageId   uint32 `field:"packageId"`   // 流量包ID
	TotalBytes  uint64 `field:"totalBytes"`  // 总字节数
	UsedBytes   uint64 `field:"usedBytes"`   // 已使用字节数
	RegionId    uint32 `field:"regionId"`    // 区域ID
	PeriodId    uint32 `field:"periodId"`    // 有效期ID
	PeriodCount uint32 `field:"periodCount"` // 有效期数量
	PeriodUnit  string `field:"periodUnit"`  // 有效期单位
	DayFrom     string `field:"dayFrom"`     // 开始日期
	DayTo       string `field:"dayTo"`       // 结束日期
	CreatedAt   uint64 `field:"createdAt"`   // 创建时间
	State       uint8  `field:"state"`       // 状态
}

type UserTrafficPackageOperator struct {
	Id          any // ID
	AdminId     any // 管理员ID
	UserId      any // 用户ID
	PackageId   any // 流量包ID
	TotalBytes  any // 总字节数
	UsedBytes   any // 已使用字节数
	RegionId    any // 区域ID
	PeriodId    any // 有效期ID
	PeriodCount any // 有效期数量
	PeriodUnit  any // 有效期单位
	DayFrom     any // 开始日期
	DayTo       any // 结束日期
	CreatedAt   any // 创建时间
	State       any // 状态
}

func NewUserTrafficPackageOperator() *UserTrafficPackageOperator {
	return &UserTrafficPackageOperator{}
}
