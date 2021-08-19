// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

const testClusterId = 7

func TestLocalEdgeDNSProvider_GetDomains(t *testing.T) {
	dbs.NotifyReady()

	provider := &dnsclients.LocalEdgeDNSProvider{}
	err := provider.Auth(maps.Map{
		"clusterId": testClusterId,
	})
	if err != nil {
		t.Fatal(err)
	}

	domains, err := provider.GetDomains()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("domains:", domains)
}

func TestLocalEdgeDNSProvider_GetRecords(t *testing.T) {
	dbs.NotifyReady()

	provider := &dnsclients.LocalEdgeDNSProvider{}
	err := provider.Auth(maps.Map{
		"clusterId": testClusterId,
	})
	if err != nil {
		t.Fatal(err)
	}

	records, err := provider.GetRecords("teaos.cn")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}

func TestLocalEdgeDNSProvider_GetRoutes(t *testing.T) {
	dbs.NotifyReady()

	provider := &dnsclients.LocalEdgeDNSProvider{}
	err := provider.Auth(maps.Map{
		"clusterId": testClusterId,
	})
	if err != nil {
		t.Fatal(err)
	}

	routes, err := provider.GetRoutes("teaos.cn")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestLocalEdgeDNSProvider_QueryRecord(t *testing.T) {
	dbs.NotifyReady()

	provider := &dnsclients.LocalEdgeDNSProvider{}
	err := provider.Auth(maps.Map{
		"clusterId": testClusterId,
	})
	if err != nil {
		t.Fatal(err)
	}
	record, err := provider.QueryRecord("teaos.cn", "cdn", dnstypes.RecordTypeA)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(record)
}

func TestLocalEdgeDNSProvider_AddRecord(t *testing.T) {
	dbs.NotifyReady()

	provider := &dnsclients.LocalEdgeDNSProvider{}
	err := provider.Auth(maps.Map{
		"clusterId": testClusterId,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = provider.AddRecord("teaos.cn", &dnstypes.Record{
		Id:    "",
		Name:  "example",
		Type:  dnstypes.RecordTypeA,
		Value: "10.0.0.1",
		Route: "id:7",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestLocalEdgeDNSProvider_UpdateRecord(t *testing.T) {
	dbs.NotifyReady()

	provider := &dnsclients.LocalEdgeDNSProvider{}
	err := provider.Auth(maps.Map{
		"clusterId": testClusterId,
	})
	if err != nil {
		t.Fatal(err)
	}

	record, err := provider.QueryRecord("teaos.cn", "cdn", dnstypes.RecordTypeA)
	if err != nil {
		t.Fatal(err)
	}
	if record == nil {
		t.Log("not found record")
		return
	}

	//record.Id = ""
	err = provider.UpdateRecord("teaos.cn", record, &dnstypes.Record{
		Id:    "",
		Name:  record.Name,
		Type:  record.Type,
		Value: "127.0.0.3",
		Route: record.Route,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestLocalEdgeDNSProvider_DeleteRecord(t *testing.T) {
	dbs.NotifyReady()

	provider := &dnsclients.LocalEdgeDNSProvider{}
	err := provider.Auth(maps.Map{
		"clusterId": testClusterId,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = provider.DeleteRecord("teaos.cn", &dnstypes.Record{
		Id:    "",
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

func TestLocalEdgeDNSProvider_DefaultRoute(t *testing.T) {
	dbs.NotifyReady()

	provider := &dnsclients.LocalEdgeDNSProvider{}
	err := provider.Auth(maps.Map{
		"clusterId": testClusterId,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(provider.DefaultRoute())
}
