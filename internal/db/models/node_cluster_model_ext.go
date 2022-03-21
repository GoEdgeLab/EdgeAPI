package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeCommon/pkg/dnsconfigs"
)

// DecodeDNSConfig 解析DNS配置
func (this *NodeCluster) DecodeDNSConfig() (*dnsconfigs.ClusterDNSConfig, error) {
	if len(this.Dns) == 0 {
		// 一定要返回一个默认的值，防止产生nil
		return &dnsconfigs.ClusterDNSConfig{
			NodesAutoSync:   false,
			ServersAutoSync: false,
		}, nil
	}
	dnsConfig := &dnsconfigs.ClusterDNSConfig{}
	err := json.Unmarshal([]byte(this.Dns), &dnsConfig)
	if err != nil {
		return nil, err
	}
	return dnsConfig, nil
}
