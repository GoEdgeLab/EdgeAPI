// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dnsclients

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestCloudFlareProvider_GetDomains(t *testing.T) {
	provider, err := testCloudFlareProvider()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(provider.GetDomains())
}

func TestCloudFlareProvider_GetRecords(t *testing.T) {
	provider, err := testCloudFlareProvider()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("===meloy.cn===")
	{
		records, err := provider.GetRecords("meloy.cn")
		if err != nil {
			t.Fatal(err)
		}
		if len(records) > 0 {
			t.Log(len(records), "records")
		}
		logs.PrintAsJSON(records, t)
	}

	t.Log("===teaos.cn===")
	{
		records, err := provider.GetRecords("teaos.cn")
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(records, t)
	}
}

func TestCloudFlareProvider_QueryRecord(t *testing.T) {
	provider, err := testCloudFlareProvider()
	if err != nil {
		t.Fatal(err)
	}
	{
		t.Log("== www.meloy.cn/A ==")
		record, err := provider.QueryRecord("meloy.cn", "www", dnstypes.RecordTypeA)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(record, t)
	}
	{
		t.Log("== www.meloy.cn/CNAME ==")
		record, err := provider.QueryRecord("meloy.cn", "www", dnstypes.RecordTypeCNAME)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(record, t)
	}
	{
		t.Log("== hello.meloy.cn ==")
		record, err := provider.QueryRecord("meloy.cn", "hello", dnstypes.RecordTypeA)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(record, t)
	}
	{
		t.Log("== test.meloy.cn ==")
		record, err := provider.QueryRecord("meloy.cn", "test", dnstypes.RecordTypeCNAME)
		if err != nil {
			t.Fatal(err)
		}
		logs.PrintAsJSON(record, t)
	}
}

func TestCloudFlareProvider_AddRecord(t *testing.T) {
	provider, err := testCloudFlareProvider()
	if err != nil {
		t.Fatal(err)
	}
	{
		err = provider.AddRecord("meloy.cn", &dnstypes.Record{
			Id:    "",
			Name:  "test",
			Type:  dnstypes.RecordTypeA,
			Value: "182.92.212.46",
			Route: "",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err = provider.AddRecord("meloy.cn", &dnstypes.Record{
			Id:    "",
			Name:  "test1",
			Type:  dnstypes.RecordTypeCNAME,
			Value: "cdn.meloy.cn.",
			Route: "",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Log("ok")
}

func TestCloudFlareProvider_UpdateRecord(t *testing.T) {
	provider, err := testCloudFlareProvider()
	if err != nil {
		t.Fatal(err)
	}
	err = provider.UpdateRecord("meloy.cn", &dnstypes.Record{Id: "b4da7ad9f90173ec37c80ba6bb70641a"}, &dnstypes.Record{
		Id:    "",
		Name:  "test1",
		Type:  dnstypes.RecordTypeCNAME,
		Value: "cdn123.meloy.cn.",
		Route: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestCloudFlareProvider_DeleteRecord(t *testing.T) {
	provider, err := testCloudFlareProvider()
	if err != nil {
		t.Fatal(err)
	}
	err = provider.DeleteRecord("meloy.cn", &dnstypes.Record{
		Id: "86282d89bbd1f66a69ca409da84f34b1",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func testCloudFlareProvider() (ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='cloudFlare' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("can not find providers with type 'cloudFlare'")
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &CloudFlareProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
