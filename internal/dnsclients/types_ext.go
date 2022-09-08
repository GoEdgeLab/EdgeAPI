// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build !plus
// +build !plus

package dnsclients

import "github.com/iwind/TeaGo/maps"

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
	case ProviderTypeCustomHTTP:
		return &CustomHTTPProvider{}
	case ProviderTypeUserEdgeDNS:
		return &EdgeDNSAPIProvider{}
	}

	return nil
}

func filterTypeMaps(typeMaps []maps.Map) []maps.Map {
	return typeMaps
}
