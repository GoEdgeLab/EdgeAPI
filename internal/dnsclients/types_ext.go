// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build !plus
// +build !plus

package dnsclients

import "github.com/iwind/TeaGo/maps"

func filterTypeMaps(typeMaps []maps.Map) []maps.Map {
	return typeMaps
}

func filterProvider(providerType string) ProviderInterface {
	return nil
}
