package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

// DecodeHTTP 解析HTTP配置
func (this *UserNode) DecodeHTTP() (*serverconfigs.HTTPProtocolConfig, error) {
	if !IsNotNull(this.Http) {
		return nil, nil
	}
	config := &serverconfigs.HTTPProtocolConfig{}
	err := json.Unmarshal([]byte(this.Http), config)
	if err != nil {
		return nil, err
	}

	err = config.Init()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// DecodeHTTPS 解析HTTPS配置
func (this *UserNode) DecodeHTTPS(cacheMap *utils.CacheMap) (*serverconfigs.HTTPSProtocolConfig, error) {
	if !IsNotNull(this.Https) {
		return nil, nil
	}
	config := &serverconfigs.HTTPSProtocolConfig{}
	err := json.Unmarshal([]byte(this.Https), config)
	if err != nil {
		return nil, err
	}

	err = config.Init()
	if err != nil {
		return nil, err
	}

	if config.SSLPolicyRef != nil {
		policyId := config.SSLPolicyRef.SSLPolicyId
		if policyId > 0 {
			sslPolicy, err := SharedSSLPolicyDAO.ComposePolicyConfig(nil, policyId, cacheMap)
			if err != nil {
				return nil, err
			}
			if sslPolicy != nil {
				config.SSLPolicy = sslPolicy
			}
		}
	}

	err = config.Init()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// DecodeAccessAddrs 解析访问地址
func (this *UserNode) DecodeAccessAddrs() ([]*serverconfigs.NetworkAddressConfig, error) {
	if !IsNotNull(this.AccessAddrs) {
		return nil, nil
	}

	addrConfigs := []*serverconfigs.NetworkAddressConfig{}
	err := json.Unmarshal([]byte(this.AccessAddrs), &addrConfigs)
	if err != nil {
		return nil, err
	}
	for _, addrConfig := range addrConfigs {
		err = addrConfig.Init()
		if err != nil {
			return nil, err
		}
	}
	return addrConfigs, nil
}

// DecodeAccessAddrStrings 解析访问地址，并返回字符串形式
func (this *UserNode) DecodeAccessAddrStrings() ([]string, error) {
	addrs, err := this.DecodeAccessAddrs()
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, addr := range addrs {
		result = append(result, addr.FullAddresses()...)
	}
	return result, nil
}
