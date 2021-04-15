package dnsclients

import "github.com/iwind/TeaGo/maps"

type ProviderType = string

// 服务商代号
const (
	ProviderTypeDNSPod     ProviderType = "dnspod"
	ProviderTypeAliDNS     ProviderType = "alidns"
	ProviderTypeDNSCom     ProviderType = "dnscom"
	ProviderTypeCloudFlare ProviderType = "cloudFlare"
	ProviderTypeCustomHTTP ProviderType = "customHTTP"
)

// AllProviderTypes 所有的服务商类型
var AllProviderTypes = []maps.Map{
	{
		"name": "阿里云DNS",
		"code": ProviderTypeAliDNS,
	},
	{
		"name": "DNSPod",
		"code": ProviderTypeDNSPod,
	},
	/**{
		"name": "帝恩思DNS.COM",
		"code": ProviderTypeDNSCom,
	},**/
	{
		"name": "CloudFlare DNS",
		"code": ProviderTypeCloudFlare,
	},
	{
		"name": "自定义HTTP DNS",
		"code": ProviderTypeCustomHTTP,
	},
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
	case ProviderTypeCustomHTTP:
		return &CustomHTTPProvider{}
	}
	return nil
}

// FindProviderTypeName 查找服务商名称
func FindProviderTypeName(providerType ProviderType) string {
	for _, t := range AllProviderTypes {
		if t.GetString("code") == providerType {
			return t.GetString("name")
		}
	}
	return ""
}
