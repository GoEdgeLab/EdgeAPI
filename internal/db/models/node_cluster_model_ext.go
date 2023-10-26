package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
)

// DecodeDNSConfig 解析DNS配置
func (this *NodeCluster) DecodeDNSConfig() (*dnsconfigs.ClusterDNSConfig, error) {
	if len(this.Dns) == 0 {
		// 一定要返回一个默认的值，防止产生nil
		return dnsconfigs.DefaultClusterDNSConfig(), nil
	}
	var dnsConfig = dnsconfigs.DefaultClusterDNSConfig()
	err := json.Unmarshal(this.Dns, &dnsConfig)
	if err != nil {
		return nil, err
	}
	return dnsConfig, nil
}

// DecodeDDoSProtection 解析DDOS Protection设置
func (this *NodeCluster) DecodeDDoSProtection() *ddosconfigs.ProtectionConfig {
	if IsNull(this.DdosProtection) {
		return nil
	}

	var result = &ddosconfigs.ProtectionConfig{}
	err := json.Unmarshal(this.DdosProtection, &result)
	if err != nil {
		// ignore err
	}
	return result
}

// HasDDoSProtection 检查是否有DDoS设置
func (this *NodeCluster) HasDDoSProtection() bool {
	var config = this.DecodeDDoSProtection()
	if config != nil {
		return config.IsOn()
	}
	return false
}

// HasNetworkSecurityPolicy 检查是否有安全策略设置
func (this *NodeCluster) HasNetworkSecurityPolicy() bool {
	var policy = this.DecodeNetworkSecurityPolicy()
	if policy != nil {
		return policy.IsOn()
	}
	return false
}

// DecodeNetworkSecurityPolicy 解析安全策略设置
func (this *NodeCluster) DecodeNetworkSecurityPolicy() *nodeconfigs.NetworkSecurityPolicy {
	var policy = nodeconfigs.NewNetworkSecurityPolicy()
	if IsNotNull(this.NetworkSecurity) {
		err := json.Unmarshal(this.NetworkSecurity, policy)
		if err != nil {
			remotelogs.Error("NodeCluster.DecodeNetworkSecurityPolicy()", err.Error())
		}
	}
	return policy
}

// DecodeClock 解析时钟配置
func (this *NodeCluster) DecodeClock() *nodeconfigs.ClockConfig {
	var clock = nodeconfigs.DefaultClockConfig()
	if IsNotNull(this.Clock) {
		err := json.Unmarshal(this.Clock, clock)
		if err != nil {
			remotelogs.Error("NodeCluster.DecodeClock()", err.Error())
		}
	}
	return clock
}

// DecodeGlobalServerConfig 解析全局服务配置
func (this *NodeCluster) DecodeGlobalServerConfig() *serverconfigs.GlobalServerConfig {
	var config = serverconfigs.NewGlobalServerConfig()
	if IsNotNull(this.GlobalServerConfig) {
		err := json.Unmarshal(this.GlobalServerConfig, config)
		if err != nil {
			remotelogs.Error("NodeCluster.DecodeGlobalServerConfig()", err.Error())
		}
	}
	return config
}
