package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/ddosconfigs"
)

// DecodeDNSConfig 解析DNS配置
func (this *NodeCluster) DecodeDNSConfig() (*dnsconfigs.ClusterDNSConfig, error) {
	if len(this.Dns) == 0 {
		// 一定要返回一个默认的值，防止产生nil
		return &dnsconfigs.ClusterDNSConfig{
			NodesAutoSync:   false,
			ServersAutoSync: false,
			CNameAsDomain:   true,
		}, nil
	}
	var dnsConfig = &dnsconfigs.ClusterDNSConfig{
		CNameAsDomain: true,
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
