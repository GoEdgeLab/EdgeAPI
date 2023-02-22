package models

// ADPackage 高防产品规格
type ADPackage struct {
	Id                      uint32 `field:"id"`                      // ID
	IsOn                    bool   `field:"isOn"`                    // 是否启用
	NetworkId               uint32 `field:"networkId"`               // 线路ID
	ProtectionBandwidthSize uint32 `field:"protectionBandwidthSize"` // 防护带宽尺寸
	ProtectionBandwidthUnit string `field:"protectionBandwidthUnit"` // 防护带宽单位
	ProtectionBandwidthBits uint64 `field:"protectionBandwidthBits"` // 防护带宽比特
	ServerBandwidthSize     uint32 `field:"serverBandwidthSize"`     // 业务带宽尺寸
	ServerBandwidthUnit     string `field:"serverBandwidthUnit"`     // 业务带宽单位
	ServerBandwidthBits     uint64 `field:"serverBandwidthBits"`     // 业务带宽比特
	State                   uint8  `field:"state"`                   // 状态
}

type ADPackageOperator struct {
	Id                      any // ID
	IsOn                    any // 是否启用
	NetworkId               any // 线路ID
	ProtectionBandwidthSize any // 防护带宽尺寸
	ProtectionBandwidthUnit any // 防护带宽单位
	ProtectionBandwidthBits any // 防护带宽比特
	ServerBandwidthSize     any // 业务带宽尺寸
	ServerBandwidthUnit     any // 业务带宽单位
	ServerBandwidthBits     any // 业务带宽比特
	State                   any // 状态
}

func NewADPackageOperator() *ADPackageOperator {
	return &ADPackageOperator{}
}
