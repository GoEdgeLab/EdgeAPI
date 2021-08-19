// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestHuaweiDNSProvider_GetDomains(t *testing.T) {
	provider, err := testHuaweiDNSProvider()
	if err != nil {
		t.Fatal(err)
	}
	domains, err := provider.GetDomains()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("domains:", domains)
}

func TestHuaweiDNSProvider_GetRecords(t *testing.T) {
	provider, err := testHuaweiDNSProvider()
	if err != nil {
		t.Fatal(err)
	}
	records, err := provider.GetRecords("yun4s.cn")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}

func TestHuaweiDNSProvider_GetRoutes(t *testing.T) {
	provider, err := testHuaweiDNSProvider()
	if err != nil {
		t.Fatal(err)
	}
	routes, err := provider.GetRoutes("yun4s.cn")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(routes))
	logs.PrintAsJSON(routes, t)
}

func TestHuaweiDNSProvider_QueryRecord(t *testing.T) {
	provider, err := testHuaweiDNSProvider()
	if err != nil {
		t.Fatal(err)
	}
	record, err := provider.QueryRecord("yun4s.cn", "abc", dnstypes.RecordTypeA)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(record)
}

func TestHuaweiDNSProvider_AddRecord(t *testing.T) {
	provider, err := testHuaweiDNSProvider()
	if err != nil {
		t.Fatal(err)
	}
	record := &dnstypes.Record{
		Id:    "",
		Name:  "add-record-1",
		Type:  "A",
		Value: "192.168.2.40",
		Route: "Beijing",
	}
	err = provider.AddRecord("yun4s.cn", record)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(record, t)
}

func TestHuaweiDNSProvider_UpdateRecord(t *testing.T) {
	provider, err := testHuaweiDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	record := &dnstypes.Record{
		Id:    "",
		Name:  "add-record-1",
		Type:  "A",
		Value: "192.168.2.42",
		Route: "default_view",
	}
	err = provider.UpdateRecord("yun4s.cn", &dnstypes.Record{
		Id: "8aace3b97ac6e108017b116f3e2e2923@192.168.2.40",
	}, record)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestHuaweiDNSProvider_DeleteRecord(t *testing.T) {
	provider, err := testHuaweiDNSProvider()
	if err != nil {
		t.Fatal(err)
	}
	record, err := provider.QueryRecord("yun4s.cn", "add-record-1", dnstypes.RecordTypeA)
	if err != nil {
		t.Fatal(err)
	}
	if record == nil {
		t.Log("not found record")
		return
	}
	err = provider.DeleteRecord("yun4s.cn", record)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func testHuaweiDNSProvider() (ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='huaweiDNS' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &HuaweiDNSProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
