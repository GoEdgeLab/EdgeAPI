package models_test

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestServerDAO_CreateManyServers(t *testing.T) {
	dbs.NotifyReady()

	var dao = models.NewServerDAO()
	var tx *dbs.Tx
	var count = 10000
	for i := 0; i < count; i++ {
		var serverNames = []*serverconfigs.ServerNameConfig{
			{
				Name: "s" + types.String(i) + ".teaos.cn",
			},
		}
		serverNamesJSON, err := json.Marshal(serverNames)
		if err != nil {
			t.Fatal(err)
		}
		serverId, err := dao.CreateServer(tx, 0, 0, serverconfigs.ServerTypeHTTPProxy, "TEST"+types.String(i), "", serverNamesJSON, false, nil, nil, nil, nil, nil, nil, nil, 0, nil, 1, nil, nil, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		_ = serverId
	}
}

func TestServerDAO_ComposeServerConfig(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := models.SharedServerDAO.ComposeServerConfigWithServerId(tx, 1, false, false)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config, t)
}

func TestServerDAO_ComposeServerConfig_AliasServerNames(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := models.SharedServerDAO.ComposeServerConfigWithServerId(tx, 14, false, false)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config.AliasServerNames, t)
}

func TestServerDAO_UpdateServerConfig(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	config, err := models.SharedServerDAO.ComposeServerConfigWithServerId(tx, 1, false, false)
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
	dnsName, err := models.SharedServerDAO.GenDNSName(tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("dnsName:", dnsName)
}

func TestServerDAO_FindAllServerDNSNamesWithDNSDomainId(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	dnsNames, err := models.SharedServerDAO.FindAllServerDNSNamesWithDNSDomainId(tx, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("dnsNames:", dnsNames)
}

func TestServerDAO_FindAllEnabledServerIdsWithSSLPolicyIds(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	serverIds, err := models.SharedServerDAO.FindAllEnabledServerIdsWithSSLPolicyIds(tx, []int64{14})
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
		isUsing, err := models.SharedServerDAO.CheckPortIsUsing(tx, 18, "tcp", 3306, 0, "")
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
		exist, err := models.SharedServerDAO.ExistServerNameInCluster(tx, 18, "hello.teaos.cn", 0, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(exist)
	}

	{
		exist, err := models.SharedServerDAO.ExistServerNameInCluster(tx, 18, "cdn.teaos.cn", 0, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(exist)
	}

	{
		exist, err := models.SharedServerDAO.ExistServerNameInCluster(tx, 18, "cdn.teaos.cn", 23, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(exist)
	}
}

func TestServerDAO_FindAllEnabledServersWithNode(t *testing.T) {
	dbs.NotifyReady()

	var before = time.Now()
	servers, err := models.SharedServerDAO.FindAllEnabledServersWithNode(nil, 48)
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
		servers, err := models.SharedServerDAO.FindAllEnabledServersWithNode(nil, 48)
		if err != nil {
			t.Fatal(err)
		}
		for _, server := range servers {
			_, _ = models.SharedServerDAO.ComposeServerConfig(nil, server, false, nil, cacheMap, true, false)
		}
	}

	var before = time.Now()
	{
		servers, err := models.SharedServerDAO.FindAllEnabledServersWithNode(nil, 48)
		if err != nil {
			t.Fatal(err)
		}
		for _, server := range servers {
			_, _ = models.SharedServerDAO.ComposeServerConfig(nil, server, false, nil, cacheMap, true, false)
		}
	}
	t.Log(time.Since(before).Seconds()*1000, "ms")
}

func TestServerDAO_FindAllEnabledServersWithDomain(t *testing.T) {
	for _, domain := range []string{"yun4s.cn", "teaos.cn", "teaos2.cn", "cdn.teaos.cn", "cdn100.teaos.cn"} {
		servers, err := models.NewServerDAO().FindAllEnabledServersWithDomain(nil, domain)
		if err != nil {
			t.Fatal(err)
		}
		if len(servers) > 0 {
			for _, server := range servers {
				t.Log(domain + ": " + string(server.ServerNames))
			}
		} else {
			t.Log(domain + ": not found")
		}
	}
}

func TestServerDAO_FindEnabledServerWithDomain(t *testing.T) {
	var dao = models.NewServerDAO()
	var tx *dbs.Tx

	for _, domain := range []string{"a", "a.com", "teaos.cn", "www.teaos.cn", "cdn.teaos.cn", "google.com"} {
		var before = time.Now()
		server, err := dao.FindEnabledServerWithDomain(tx, 0, domain)
		var costMs = time.Since(before).Seconds() * 1000
		if err != nil {
			t.Fatal(err)
		}
		if server == nil {
			t.Log(domain, "NULL", fmt.Sprintf("%.2fms", costMs))
		} else {
			t.Log(domain, string(maps.Map{"id": server.Id, "clusterId": server.ClusterId, "userId": server.UserId}.AsJSON()), fmt.Sprintf("%.2fms", costMs))
		}
	}
}

func TestServerDAO_RenewServerTrafficLimitStatus(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	before := time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	err := models.NewServerDAO().RenewServerTrafficLimitStatus(tx, &serverconfigs.TrafficLimitConfig{
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

func TestServerDAO_UpdateServerTrafficLimitStatus(t *testing.T) {
	dbs.NotifyReady()

	var dao = models.NewServerDAO()
	var tx *dbs.Tx
	err := dao.UpdateServerTrafficLimitStatus(tx, 23, timeutil.Format("Ymd", time.Now().AddDate(0, 0, 20)), 14, "day")
	if err != nil {
		t.Fatal(err)
	}
}

func TestServerDAO_FindBytes(t *testing.T) {
	col, err := models.NewServerDAO().Query(nil).
		Result("http").
		Pk(1).
		FindBytesCol()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(col))
}

func TestServerDAO_FindBool(t *testing.T) {
	one, err := models.NewServerDAO().Query(nil).
		Result("isOn").
		Pk(1).
		Find()
	if err != nil {
		t.Fatal(err)
	}
	if one != nil {
		t.Log(one.(*models.Server).IsOn)
	}
}

func TestServerDAO_UpdateServerBandwidth(t *testing.T) {
	var dao = models.NewServerDAO()
	var tx *dbs.Tx
	err := dao.UpdateServerBandwidth(tx, 1, timeutil.FormatTime("YmdHi", time.Now().Unix()/300*300), 1024, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServerDAO_FindEnabledServersWithIds(t *testing.T) {
	var dao = models.NewServerDAO()
	var tx *dbs.Tx
	servers, err := dao.FindEnabledServersWithIds(tx, []int64{23, 1071})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(servers), "servers")
}

func BenchmarkServerDAO_CountAllEnabledServers(b *testing.B) {
	models.SharedServerDAO = models.NewServerDAO()

	for i := 0; i < b.N; i++ {
		result, err := models.SharedServerDAO.CountAllEnabledServers(nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}
