package dnsclients

import "github.com/iwind/TeaGo/maps"

type ProviderType = string

// 服务商代号
const (
	ProviderTypeDNSPod ProviderType = "dnspod"
	ProviderTypeAliyun ProviderType = "aliyun"
	ProviderTypeDNSCom ProviderType = "dnscom"
)

// 所有的服务商类型
var AllProviderTypes = []maps.Map{
	{
		"name": "DNSPod",
		"code": ProviderTypeDNSPod,
	},
	{
		"name": "阿里云",
		"code": ProviderTypeAliyun,
	},
	{
		"name": "帝恩思",
		"code": ProviderTypeDNSCom,
	},
}

// 查找服务商实例
func FindProvider(providerType ProviderType) ProviderInterface {
	switch providerType {
	case ProviderTypeDNSPod:
		return &DNSPodProvider{}
	}
	return nil
}

// 查找服务商名称
func FindProviderTypeName(providerType ProviderType) string {
	for _, t := range AllProviderTypes {
		if t.GetString("code") == providerType {
			return t.GetString("name")
		}
	}
	return ""
}
