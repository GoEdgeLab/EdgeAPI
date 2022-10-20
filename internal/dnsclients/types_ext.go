// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build !plus
// +build !plus

package dnsclients

import "github.com/iwind/TeaGo/maps"

// FindProvider 查找服务商实例
func FindProvider(providerType ProviderType, providerId int64) ProviderInterface {
	switch providerType {
	case ProviderTypeDNSPod:
		return &DNSPodProvider{
			ProviderId: providerId,
		}
	case ProviderTypeAliDNS:
		return &AliDNSProvider{
			ProviderId: providerId,
		}
	case ProviderTypeHuaweiDNS:
		return &HuaweiDNSProvider{
			ProviderId: providerId,
		}
	case ProviderTypeCloudFlare:
		return &CloudFlareProvider{
			ProviderId: providerId,
		}
	case ProviderTypeCustomHTTP:
		return &CustomHTTPProvider{
			ProviderId: providerId,
		}
	case ProviderTypeEdgeDNSAPI:
		return &EdgeDNSAPIProvider{
			ProviderId: providerId,
		}
	}

	return nil
}

func filterTypeMaps(typeMaps []maps.Map) []maps.Map {
	return typeMaps
}
