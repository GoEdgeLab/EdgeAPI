package dnsclients

import (
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/iwind/TeaGo/maps"
)

type ProviderType = string

// 服务商代号
const (
	ProviderTypeDNSPod       ProviderType = "dnspod"       // DNSPod
	ProviderTypeAliDNS       ProviderType = "alidns"       // 阿里云DNS
	ProviderTypeDNSCom       ProviderType = "dnscom"       // dns.com
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
		/**{
			"name": "帝恩思DNS.COM",
			"code": ProviderTypeDNSCom,
			"description": "DNS.com提供的DNS服务。",
		},**/
		{
			"name":        "CloudFlare DNS",
			"code":        ProviderTypeCloudFlare,
			"description": "CloudFlare提供的DNS服务。",
		},
	}

	if teaconst.IsPlus {
		typeMaps = append(typeMaps, []maps.Map{
			{
				"name":        "自建EdgeDNS",
				"code":        ProviderTypeLocalEdgeDNS,
				"description": "当前企业版提供的自建DNS服务。",
			},
			// TODO 需要实现用户使用AccessId/AccessKey来连接DNS服务
			/**{
				"name":        "用户EdgeDNS",
				"code":        ProviderTypeUserEdgeDNS,
				"description": "通过API连接企业版提供的DNS服务。",
			},**/
		}...)
	}

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
	case ProviderTypeCloudFlare:
		return &CloudFlareProvider{}
	case ProviderTypeLocalEdgeDNS:
		return &LocalEdgeDNSProvider{}
	case ProviderTypeUserEdgeDNS:
		return &UserEdgeDNSProvider{}
	case ProviderTypeCustomHTTP:
		return &CustomHTTPProvider{}
	}
	return nil
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
