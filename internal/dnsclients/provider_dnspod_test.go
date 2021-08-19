package dnsclients

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestDNSPodProvider_GetDomains(t *testing.T) {
	provider, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}
	domains, err := provider.GetDomains()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(domains)
}

func TestDNSPodProvider_GetRoutes(t *testing.T) {
	provider, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}
	routes, err := provider.GetRoutes("yun4s.cn")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestDNSPodProvider_GetRecords(t *testing.T) {
	provider, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}
	records, err := provider.GetRecords("yun4s.cn")
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		t.Log(record.Id, record.Type, record.Name, record.Value, record.Route)
	}
}

func TestDNSPodProvider_AddRecord(t *testing.T) {
	provider, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.AddRecord("yun4s.cn", &dnstypes.Record{
		Type:  dnstypes.RecordTypeCNAME,
		Name:  "hello-forward",
		Value: "hello.yun4s.cn",
		Route: "联通",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestDNSPodProvider_UpdateRecord(t *testing.T) {
	provider, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.UpdateRecord("yun4s.cn", &dnstypes.Record{
		Id: "697036856",
	}, &dnstypes.Record{
		Type:  dnstypes.RecordTypeA,
		Name:  "hello",
		Value: "192.168.1.102",
		Route: "联通",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestDNSPodProvider_DeleteRecord(t *testing.T) {
	provider, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.DeleteRecord("yun4s.cn", &dnstypes.Record{
		Id: "697040986",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func testDNSPodProvider() (ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='dnspod' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &DNSPodProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
