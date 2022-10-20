// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dnsclients_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestDomainRecordsCache_WriteDomainRecords(t *testing.T) {
	dbs.NotifyReady()

	var cache = dnsclients.NewDomainRecordsCache()
	cache.WriteDomainRecords(1, "a", []*dnstypes.Record{
		{
			Id:    "1",
			Name:  "hello",
			Type:  "A",
			Value: "192.168.1.100",
		},
	})

	//time.Sleep(30 * time.Second)

	{
		t.Log(cache.QueryDomainRecord(1, "a", "hello", "A"))
	}
	{
		t.Log(cache.QueryDomainRecord(1, "a", "hello", "AAAA"))
	}
	{
		t.Log(cache.QueryDomainRecord(1, "a", "hello2", "A"))
	}

	t.Log("======")
	cache.DeleteDomainRecord(1, "a", "2")
	cache.UpdateDomainRecord(1, "a", &dnstypes.Record{
		Id:    "1",
		Name:  "hello2",
		Type:  "A",
		Value: "192.168.1.200",
	})
	{
		t.Log(cache.QueryDomainRecord(1, "a", "hello2", "A"))
	}
	t.Log("======")
	cache.AddDomainRecord(1, "a", &dnstypes.Record{
		Id:    "2",
		Name:  "hello",
		Type:  "AAAA",
		Value: "::1",
	})
	{
		t.Log(cache.QueryDomainRecord(1, "a", "hello", "AAAA"))
	}
}
