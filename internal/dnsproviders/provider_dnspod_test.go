package dnsproviders

import (
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestDNSPodProvider_GetRoutes(t *testing.T) {
	provider := &DNSPodProvider{}
	err := provider.Auth(maps.Map{
		"id":    "191996",
		"token": "366964e0f8ed4d8990a7f5d4b3cdec60",
	})
	if err != nil {
		t.Fatal(err)
	}
	routes, err := provider.GetRoutes("yun4s.cn")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(routes)
}
