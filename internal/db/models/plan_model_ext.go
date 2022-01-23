package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

// DecodeTrafficPrice 流量价格配置
func (this *Plan) DecodeTrafficPrice() *serverconfigs.PlanTrafficPriceConfig {
	var config = &serverconfigs.PlanTrafficPriceConfig{}

	if len(this.TrafficPrice) == 0 {
		return config
	}

	err := json.Unmarshal([]byte(this.TrafficPrice), config)
	if err != nil {
		// 忽略错误
	}

	return config
}

// DecodeBandwidthPrice 带宽价格配置
func (this *Plan) DecodeBandwidthPrice() *serverconfigs.PlanBandwidthPriceConfig {
	var config = &serverconfigs.PlanBandwidthPriceConfig{}

	if len(this.BandwidthPrice) == 0 {
		return config
	}

	err := json.Unmarshal([]byte(this.BandwidthPrice), config)
	if err != nil {
		// 忽略错误
	}

	return config
}
