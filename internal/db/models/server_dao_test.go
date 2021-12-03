package models

import (
	"crypto/md5"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"testing"
	"time"
)

func TestServerDAO_ComposeServerConfig(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := SharedServerDAO.ComposeServerConfigWithServerId(tx, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config, t)
}

func TestServerDAO_ComposeServerConfig_AliasServerNames(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := SharedServerDAO.ComposeServerConfigWithServerId(tx, 14, false)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config.AliasServerNames, t)
}

func TestServerDAO_UpdateServerConfig(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := SharedServerDAO.ComposeServerConfigWithServerId(tx, 1, false)
	if err != nil {
		t.Fatal(err)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(configJSON))
	t.Log("ok")
}

func TestNewServerDAO_md5(t *testing.T) {
	m := md5.New()
	_, err := m.Write([]byte("123456"))
	if err != nil {
		t.Fatal(err)
	}
	h := m.Sum(nil)
	t.Logf("%x", h)
}

func TestServerDAO_genDNSName(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	dnsName, err := SharedServerDAO.GenDNSName(tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("dnsName:", dnsName)
}

func TestServerDAO_FindAllServerDNSNamesWithDNSDomainId(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	dnsNames, err := SharedServerDAO.FindAllServerDNSNamesWithDNSDomainId(tx, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("dnsNames:", dnsNames)
}

func TestServerDAO_FindAllEnabledServerIdsWithSSLPolicyIds(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	serverIds, err := SharedServerDAO.FindAllEnabledServerIdsWithSSLPolicyIds(tx, []int64{14})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("serverIds:", serverIds)
}

func TestServerDAO_CheckPortIsUsing(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	//{
	//	isUsing, err := SharedServerDAO.CheckPortIsUsing(tx, 18, 1234, 0, "")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	t.Log("isUsing:", isUsing)
	//}
	{
		isUsing, err := SharedServerDAO.CheckPortIsUsing(tx, 18, "tcp", 3306, 0, "")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("isUsing:", isUsing)
	}
}

func TestServerDAO_ExistServerNameInCluster(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	{
		exist, err := SharedServerDAO.ExistServerNameInCluster(tx, 18, "hello.teaos.cn", 0)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(exist)
	}

	{
		exist, err := SharedServerDAO.ExistServerNameInCluster(tx, 18, "cdn.teaos.cn", 0)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(exist)
	}

	{
		exist, err := SharedServerDAO.ExistServerNameInCluster(tx, 18, "cdn.teaos.cn", 23)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(exist)
	}
}

func TestServerDAO_FindAllEnabledServersWithNode(t *testing.T) {
	dbs.NotifyReady()

	var before = time.Now()
	servers, err := SharedServerDAO.FindAllEnabledServersWithNode(nil, 48)
	if err != nil {
		t.Fatal(err)
	}
	for _, server := range servers {
		t.Log("serverId:", server.Id, "clusterId:", server.ClusterId)
	}
	t.Log(time.Since(before).Seconds()*1000, "ms")
}

func TestServerDAO_FindAllEnabledServersWithNode_Cache(t *testing.T) {
	dbs.NotifyReady()

	var cacheMap = utils.NewCacheMap()
	{
		servers, err := SharedServerDAO.FindAllEnabledServersWithNode(nil, 48)
		if err != nil {
			t.Fatal(err)
		}
		for _, server := range servers {
			_, _ = SharedServerDAO.ComposeServerConfig(nil, server, cacheMap, true)
		}
	}

	var before = time.Now()
	{
		servers, err := SharedServerDAO.FindAllEnabledServersWithNode(nil, 48)
		if err != nil {
			t.Fatal(err)
		}
		for _, server := range servers {
			_, _ = SharedServerDAO.ComposeServerConfig(nil, server, cacheMap, true)
		}
	}
	t.Log(time.Since(before).Seconds()*1000, "ms")
}

func TestServerDAO_FindAllEnabledServersWithDomain(t *testing.T) {
	for _, domain := range []string{"yun4s.cn", "teaos.cn", "teaos2.cn", "cdn.teaos.cn", "cdn100.teaos.cn"} {
		servers, err := NewServerDAO().FindAllEnabledServersWithDomain(nil, domain)
		if err != nil {
			t.Fatal(err)
		}
		if len(servers) > 0 {
			for _, server := range servers {
				t.Log(domain + ": " + server.ServerNames)
			}
		} else {
			t.Log(domain + ": not found")
		}
	}
}

func TestServerDAO_UpdateServerTrafficLimitStatus(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	before := time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	err := NewServerDAO().UpdateServerTrafficLimitStatus(tx, &serverconfigs.TrafficLimitConfig{
		IsOn:           true,
		DailySize:      &shared.SizeCapacity{Count: 1, Unit: "mb"},
		MonthlySize:    &shared.SizeCapacity{Count: 10, Unit: "mb"},
		TotalSize:      &shared.SizeCapacity{Count: 100, Unit: "gb"},
		NoticePageBody: "",
	}, 23, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestServerDAO_CalculateServerTrafficLimitConfig(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	before := time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()

	var cacheMap = utils.NewCacheMap()
	config, err := SharedServerDAO.CalculateServerTrafficLimitConfig(tx, 23, cacheMap)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config, t)
}

func TestServerDAO_CalculateServerTrafficLimitConfig_Cache(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	before := time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()

	var cacheMap = utils.NewCacheMap()
	for i := 0; i < 10; i++ {
		config, err := SharedServerDAO.CalculateServerTrafficLimitConfig(tx, 23, cacheMap)
		if err != nil {
			t.Fatal(err)
		}
		_ = config
	}
}

func BenchmarkServerDAO_CountAllEnabledServers(b *testing.B) {
	SharedServerDAO = NewServerDAO()

	for i := 0; i < b.N; i++ {
		result, err := SharedServerDAO.CountAllEnabledServers(nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}
