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
		return &dnsconfigs.ClusterDNSConfig{
			NodesAutoSync:    false,
			ServersAutoSync:  false,
			CNameAsDomain:    true,
			IncludingLnNodes: true,
		}, nil
	}
	var dnsConfig = &dnsconfigs.ClusterDNSConfig{
		CNameAsDomain:    true,
		IncludingLnNodes: true,
	}
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

// HasDDoSProtection 检查是否有DDOS设置
func (this *NodeCluster) HasDDoSProtection() bool {
	var config = this.DecodeDDoSProtection()
	if config != nil {
		return config.IsOn()
	}
	return false
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
	var config = serverconfigs.DefaultGlobalServerConfig()
	if IsNotNull(this.GlobalServerConfig) {
		err := json.Unmarshal(this.GlobalServerConfig, config)
		if err != nil {
			remotelogs.Error("NodeCluster.DecodeGlobalServerConfig()", err.Error())
		}
	}
	return config
}
