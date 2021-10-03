// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package acme

const DefaultProviderCode = "letsencrypt"

type Provider struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	APIURL         string `json:"apiURL"`
	RequireEAB     bool   `json:"requireEAB"`
	EABDescription string `json:"eabDescription"`
}

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

func FindProviderWithCode(code string) *Provider {
	for _, provider := range FindAllProviders() {
		if provider.Code == code {
			return provider
		}
	}
	return nil
}
