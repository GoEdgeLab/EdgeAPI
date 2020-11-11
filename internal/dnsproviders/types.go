package dnsproviders

import "github.com/iwind/TeaGo/maps"

type ProviderType = string

const (
	ProviderTypeDNSPod ProviderType = "dnspod"
)

var AllProviderTypes = []maps.Map{
	{
		"name": "DNSPod",
		"code": ProviderTypeDNSPod,
	},
}

func FindProviderTypeName(providerType string) string {
	for _, t := range AllProviderTypes {
		if t.GetString("code") == providerType {
			return t.GetString("name")
		}
	}
	return ""
}
