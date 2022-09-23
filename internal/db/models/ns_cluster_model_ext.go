package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
)

// DecodeDDoSProtection 解析DDOS Protection设置
func (this *NSCluster) DecodeDDoSProtection() *ddosconfigs.ProtectionConfig {
	if IsNull(this.DdosProtection) {
		return nil
	}

	var result = &ddosconfigs.ProtectionConfig{}
	err := json.Unmarshal(this.DdosProtection, &result)
	if err != nil {
		remotelogs.Error("NSCluster.DecodeDDoSProtection", "decode failed: "+err.Error())
	}
	return result
}

// HasDDoSProtection 检查是否有DDOS设置
func (this *NSCluster) HasDDoSProtection() bool {
	var config = this.DecodeDDoSProtection()
	if config != nil {
		return config.IsOn()
	}
	return false
}

// DecodeHosts 解析主机地址
func (this *NSCluster) DecodeHosts() []string {
	if IsNull(this.Hosts) {
		return nil
	}

	var hosts = []string{}
	err := json.Unmarshal(this.Hosts, &hosts)
	if err != nil {
		remotelogs.Error("NSCluster.DecodeHosts", "decode failed: "+err.Error())
	}

	return hosts
}

// DecodeAnswerConfig 解析应答设置
func (this *NSCluster) DecodeAnswerConfig() *dnsconfigs.NSAnswerConfig {
	var config = dnsconfigs.DefaultNSAnswerConfig()

	if IsNull(this.Answer) {
		return config
	}

	err := json.Unmarshal(this.Answer, config)
	if err != nil {
		remotelogs.Error("NSCluster.DecodeAnswerConfig", "decode failed: "+err.Error())
	}

	return config
}
