package dnsclients

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients/dnstypes"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"testing"
)

func TestAliDNSProvider_GetDomains(t *testing.T) {
	provider, err := testAliDNSProvider()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(provider.GetDomains())
}

func TestAliDNSProvider_GetRecords(t *testing.T) {
	provider, err := testAliDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	records, err := provider.GetRecords("meloy.cn")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(records, t)
}

func TestAliDNSProvider_DeleteRecord(t *testing.T) {
	provider, err := testAliDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.DeleteRecord("meloy.cn", &dnstypes.Record{
		Id: "20746603318032384",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestAliDNSProvider_GetRoutes(t *testing.T) {
	provider, err := testAliDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	routes, err := provider.GetRoutes("meloy.cn")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(routes, t)
}

func TestAliDNSProvider_AddRecord(t *testing.T) {
	provider, err := testAliDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.AddRecord("meloy.cn", &dnstypes.Record{
		Id:    "",
		Name:  "test",
		Type:  dnstypes.RecordTypeA,
		Value: "192.168.1.100",
		Route: "aliyun_r_cn-beijing",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestAliDNSProvider_UpdateRecord(t *testing.T) {
	provider, err := testAliDNSProvider()
	if err != nil {
		t.Fatal(err)
	}

	err = provider.UpdateRecord("meloy.cn", &dnstypes.Record{Id: "20746664455255040"}, &dnstypes.Record{
		Id:    "",
		Name:  "test",
		Type:  dnstypes.RecordTypeA,
		Value: "192.168.1.101",
		Route: "unicom",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func testAliDNSProvider() (ProviderInterface, error) {
	db, err := dbs.Default()
	if err != nil {
		return nil, err
	}
	one, err := db.FindOne("SELECT * FROM edgeDNSProviders WHERE type='alidns' AND state=1 ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	apiParams := maps.Map{}
	err = json.Unmarshal([]byte(one.GetString("apiParams")), &apiParams)
	if err != nil {
		return nil, err
	}
	provider := &AliDNSProvider{}
	err = provider.Auth(apiParams)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
