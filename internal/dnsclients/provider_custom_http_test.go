package dnsclients

import (
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestCustomHTTPProvider_AddRecord(t *testing.T) {
	provider := CustomHTTPProvider{}
	err := provider.Auth(maps.Map{
		"url":    "http://127.0.0.1:1234/dns",
		"secret": "123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = provider.AddRecord("hello.com", &dnstypes.Record{
		Id:    "",
		Name:  "world",
		Type:  dnstypes.RecordTypeA,
		Value: "127.0.0.1",
		Route: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestCustomHTTPProvider_GetRecords(t *testing.T) {
	provider := CustomHTTPProvider{}
	err := provider.Auth(maps.Map{
		"url":    "http://127.0.0.1:1234/dns",
		"secret": "123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	records, err := provider.GetRecords("hello.com")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}
