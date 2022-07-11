package dnsclients

import (
	"github.com/iwind/TeaGo/maps"
)

type ProviderType = string

// 服务商代号
const (
	ProviderTypeDNSPod       ProviderType = "dnspod"       // DNSPod
	ProviderTypeAliDNS       ProviderType = "alidns"       // 阿里云DNS
	ProviderTypeHuaweiDNS    ProviderType = "huaweiDNS"    // 华为DNS
	ProviderTypeCloudFlare   ProviderType = "cloudFlare"   // CloudFlare DNS
	ProviderTypeLocalEdgeDNS ProviderType = "localEdgeDNS" // 和当前系统集成的EdgeDNS
	ProviderTypeUserEdgeDNS  ProviderType = "userEdgeDNS"  // 通过API连接的EdgeDNS
	ProviderTypeCustomHTTP   ProviderType = "customHTTP"   // 自定义HTTP接口
)

// FindAllProviderTypes 所有的服务商类型
func FindAllProviderTypes() []maps.Map {
	typeMaps := []maps.Map{
		{
			"name":        "阿里云DNS",
			"code":        ProviderTypeAliDNS,
			"description": "阿里云提供的DNS服务。",
		},
		{
			"name":        "DNSPod",
			"code":        ProviderTypeDNSPod,
			"description": "DNSPod提供的DNS服务。",
		},
		{
			"name":        "华为云DNS",
			"code":        ProviderTypeHuaweiDNS,
			"description": "华为云解析DNS。",
		},
		{
			"name":        "CloudFlare DNS",
			"code":        ProviderTypeCloudFlare,
			"description": "CloudFlare提供的DNS服务。",
		},
	}

	typeMaps = filterTypeMaps(typeMaps)

	typeMaps = append(typeMaps, maps.Map{
		"name":        "自定义HTTP DNS",
		"code":        ProviderTypeCustomHTTP,
		"description": "通过自定义的HTTP接口提供DNS服务。",
	})
	return typeMaps
}

// FindProvider 查找服务商实例
func FindProvider(providerType ProviderType) ProviderInterface {
	switch providerType {
	case ProviderTypeDNSPod:
		return &DNSPodProvider{}
	case ProviderTypeAliDNS:
		return &AliDNSProvider{}
	case ProviderTypeHuaweiDNS:
		return &HuaweiDNSProvider{}
	case ProviderTypeCloudFlare:
		return &CloudFlareProvider{}
	case ProviderTypeLocalEdgeDNS:
		return &LocalEdgeDNSProvider{}
	case ProviderTypeUserEdgeDNS:
		return &UserEdgeDNSProvider{}
	case ProviderTypeCustomHTTP:
		return &CustomHTTPProvider{}
	}

	return filterProvider(providerType)
}

// FindProviderTypeName 查找服务商名称
func FindProviderTypeName(providerType ProviderType) string {
	for _, t := range FindAllProviderTypes() {
		if t.GetString("code") == providerType {
			return t.GetString("name")
		}
	}
	return ""
}
