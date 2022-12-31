package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/dbs"
)

// DecodeHTTP 解析HTTP配置
func (this *APINode) DecodeHTTP() (*serverconfigs.HTTPProtocolConfig, error) {
	if !IsNotNull(this.Http) {
		return nil, nil
	}
	config := &serverconfigs.HTTPProtocolConfig{}
	err := json.Unmarshal(this.Http, config)
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
func (this *APINode) DecodeHTTPS(tx *dbs.Tx, cacheMap *utils.CacheMap) (*serverconfigs.HTTPSProtocolConfig, error) {
	if !IsNotNull(this.Https) {
		return nil, nil
	}
	config := &serverconfigs.HTTPSProtocolConfig{}
	err := json.Unmarshal(this.Https, config)
	if err != nil {
		return nil, err
	}

	err = config.Init()
	if err != nil {
		return nil, err
	}

	if config.SSLPolicyRef != nil {
		var policyId = config.SSLPolicyRef.SSLPolicyId
		if policyId > 0 {
			sslPolicy, err := SharedSSLPolicyDAO.ComposePolicyConfig(tx, policyId, false, cacheMap)
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
func (this *APINode) DecodeAccessAddrs() ([]*serverconfigs.NetworkAddressConfig, error) {
	if !IsNotNull(this.AccessAddrs) {
		return nil, nil
	}

	addrConfigs := []*serverconfigs.NetworkAddressConfig{}
	err := json.Unmarshal(this.AccessAddrs, &addrConfigs)
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
func (this *APINode) DecodeAccessAddrStrings() ([]string, error) {
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

// DecodeRestHTTP 解析Rest HTTP配置
func (this *APINode) DecodeRestHTTP() (*serverconfigs.HTTPProtocolConfig, error) {
	if this.RestIsOn != 1 {
		return nil, nil
	}
	if !IsNotNull(this.RestHTTP) {
		return nil, nil
	}
	config := &serverconfigs.HTTPProtocolConfig{}
	err := json.Unmarshal(this.RestHTTP, config)
	if err != nil {
		return nil, err
	}

	err = config.Init()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// DecodeRestHTTPS 解析HTTPS配置
func (this *APINode) DecodeRestHTTPS(tx *dbs.Tx, cacheMap *utils.CacheMap) (*serverconfigs.HTTPSProtocolConfig, error) {
	if cacheMap == nil {
		cacheMap = utils.NewCacheMap()
	}
	if this.RestIsOn != 1 {
		return nil, nil
	}
	if !IsNotNull(this.RestHTTPS) {
		return nil, nil
	}
	config := &serverconfigs.HTTPSProtocolConfig{}
	err := json.Unmarshal(this.RestHTTPS, config)
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
			sslPolicy, err := SharedSSLPolicyDAO.ComposePolicyConfig(tx, policyId, false, cacheMap)
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
