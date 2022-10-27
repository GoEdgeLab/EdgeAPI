// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package acme

func FindAllProviders() []*Provider {
	return []*Provider{
		{
			Name:        "Let's Encrypt",
			Code:        DefaultProviderCode,
			Description: "非盈利组织Let's Encrypt提供的免费证书。",
			APIURL:      "https://acme-v02.api.letsencrypt.org/directory",
			RequireEAB:  false,
		},
		{
			Name:           "ZeroSSL",
			Code:           "zerossl",
			Description:    "相关文档 <a href=\"https://zerossl.com/documentation/acme/\" target=\"_blank\">https://zerossl.com/documentation/acme/</a>。",
			APIURL:         "https://acme.zerossl.com/v2/DV90",
			RequireEAB:     true,
			EABDescription: "在官网<a href=\"https://app.zerossl.com/developer\" target=\"_blank\">[Developer]</a>页面底部点击\"Generate\"按钮生成。",
		},
	}
}
