package dnsclients_test

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

const DNSPodTestDomain = "yun4s.cn"

func TestDNSPodProvider_GetDomains(t *testing.T) {
	provider, _, err := testDNSPodProvider()
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
	provider, _, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}
	routes, err := provider.GetRoutes(DNSPodTestDomain)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestDNSPodProvider_GetRecords(t *testing.T) {
	provider, _, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}
	records, err := provider.GetRecords(DNSPodTestDomain)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		t.Log(record.Id, record.Type, record.Name, record.Value, record.Route)
	}
}

func TestDNSPodProvider_AddRecord(t *testing.T) {
	provider, isInternational, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}

	var route = "联通"
	if isInternational {
		route = "Default"
	}

	err = provider.AddRecord(DNSPodTestDomain, &dnstypes.Record{
		Type:  dnstypes.RecordTypeCNAME,
		Name:  "hello-forward",
		Value: "hello." + DNSPodTestDomain,
		Route: route,
		TTL:   600,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestDNSPodProvider_UpdateRecord(t *testing.T) {
	provider, isInternational, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}

	var route = "联通"
	var id = "1093875360"
	if isInternational {
		route = "Default"
		id = "28507333"
	}

	err = provider.UpdateRecord(DNSPodTestDomain, &dnstypes.Record{
		Id: id,
	}, &dnstypes.Record{
		Type:  dnstypes.RecordTypeA,
		Name:  "hello",
		Value: "192.168.1.102",
		Route: route,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestDNSPodProvider_DeleteRecord(t *testing.T) {
	provider, isInternational, err := testDNSPodProvider()
	if err != nil {
		t.Fatal(err)
	}

	var id = "1093875360"
	if isInternational {
		id = "28507333"
	}
	err = provider.DeleteRecord(DNSPodTestDomain, &dnstypes.Record{
		Id: id,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func testDNSPodProvider() (provider dnsclients.ProviderInterface, isInternational bool, err error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, false, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='dnspod' AND id='14' ORDER BY id DESC")
	if err != nil {
		return nil, false, err
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, false, err
	}
	provider = &dnsclients.DNSPodProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, false, err
	}
	return provider, apiParams.GetString("region") == "international", nil
}
