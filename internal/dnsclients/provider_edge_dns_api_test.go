// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

const edgeDNSAPIDomainName = "hello2.com"

func TestEdgeDNSAPIProvider_GetDomains(t *testing.T) {
	provider, err := testEdgeDNSAPIProvider()
	if err != nil {
		t.Fatal(err)
	}

	domains, err := provider.GetDomains()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("domains:", domains)
}

func TestEdgeDNSAPIProvider_GetRecords(t *testing.T) {
	provider, err := testEdgeDNSAPIProvider()
	if err != nil {
		t.Fatal(err)
	}

	records, err := provider.GetRecords(edgeDNSAPIDomainName)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}

func TestEdgeDNSAPIProvider_GetRoutes(t *testing.T) {
	provider, err := testEdgeDNSAPIProvider()
	if err != nil {
		t.Fatal(err)
	}

	routes, err := provider.GetRoutes(edgeDNSAPIDomainName)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestEdgeDNSAPIProvider_QueryRecord(t *testing.T) {
	provider, err := testEdgeDNSAPIProvider()
	if err != nil {
		t.Fatal(err)
	}
	record, err := provider.QueryRecord(edgeDNSAPIDomainName, "cdn", dnstypes.RecordTypeA)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(record)
}

func TestEdgeDNSAPIProvider_AddRecord(t *testing.T) {
	provider, err := testEdgeDNSAPIProvider()
	if err != nil {
		t.Fatal(err)
	}
	err = provider.AddRecord(edgeDNSAPIDomainName, &dnstypes.Record{
		Id:    "",
		Name:  "example",
		Type:  dnstypes.RecordTypeA,
		Value: "10.0.0.1",
		Route: "china:province:beijing",
		TTL:   300,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestEdgeDNSAPIProvider_UpdateRecord(t *testing.T) {
	provider, err := testEdgeDNSAPIProvider()
	if err != nil {
		t.Fatal(err)
	}
	record, err := provider.QueryRecord(edgeDNSAPIDomainName, "cdn", dnstypes.RecordTypeA)
	if err != nil {
		t.Fatal(err)
	}
	if record == nil {
		t.Log("not found record")
		return
	}

	//record.Id = ""
	err = provider.UpdateRecord(edgeDNSAPIDomainName, record, &dnstypes.Record{
		Id:    "",
		Name:  record.Name,
		Type:  record.Type,
		Value: "127.0.0.3",
		Route: record.Route,
		TTL:   30,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestEdgeDNSAPIProvider_DeleteRecord(t *testing.T) {
	provider, err := testEdgeDNSAPIProvider()
	if err != nil {
		t.Fatal(err)
	}

	record, err := provider.QueryRecord(edgeDNSAPIDomainName, "example", "A")
	if err != nil {
		t.Fatal(err)
	}
	if record == nil {
		t.Log("not found")
		return
	}

	err = provider.DeleteRecord(edgeDNSAPIDomainName, &dnstypes.Record{
		Id:    record.Id,
		Name:  "example",
		Type:  "A",
		Value: "",
		Route: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestEdgeDNSAPIProvider_DefaultRoute(t *testing.T) {
	provider, err := testEdgeDNSAPIProvider()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(provider.DefaultRoute())
}

func testEdgeDNSAPIProvider() (dnsclients.ProviderInterface, error) {
	provider := &dnsclients.EdgeDNSAPIProvider{}
	err := provider.Auth(maps.Map{
		"role":            "user",
		"host":            "http://127.0.0.1:8004",
		"accessKeyId":     "JOvsyXIFqkQbh5kl",
		"accessKeySecret": "t0RY8YO3R58VbJJNp0RqKw9KWNpObwtE",
	})
	if err != nil {
		return nil, err
	}
	return provider, nil
}
