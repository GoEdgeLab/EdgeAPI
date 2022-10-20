package models

// TrafficPackage 流量包
type TrafficPackage struct {
	Id    uint32 `field:"id"`    // ID
	Size  uint32 `field:"size"`  // 尺寸
	Unit  string `field:"unit"`  // 单位（gb|tb等）
	Bytes uint64 `field:"bytes"` // 字节
	IsOn  bool   `field:"isOn"`  // 是否启用
	State uint8  `field:"state"` // 状态
}

type TrafficPackageOperator struct {
	Id    any // ID
	Size  any // 尺寸
	Unit  any // 单位（gb|tb等）
	Bytes any // 字节
	IsOn  any // 是否启用
	State any // 状态
}

func NewTrafficPackageOperator() *TrafficPackageOperator {
	return &TrafficPackageOperator{}
}
