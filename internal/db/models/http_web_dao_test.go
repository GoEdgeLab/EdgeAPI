package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHTTPWebDAO_UpdateWebShutdown(t *testing.T) {
	var tx *dbs.Tx
	{
		err := SharedHTTPWebDAO.UpdateWebShutdown(tx, 1, []byte("{}"))
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		err := SharedHTTPWebDAO.UpdateWebShutdown(tx, 1, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Log("ok")
}

func TestHTTPWebDAO_FindAllWebIdsWithHTTPFirewallPolicyId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx

	webIds, err := SharedHTTPWebDAO.FindAllWebIdsWithHTTPFirewallPolicyId(tx, 9)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("webIds:", webIds)

	count, err := SharedServerDAO.CountEnabledServersWithWebIds(tx, webIds)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("count:", count)
}

func TestHTTPWebDAO_FindWebServerId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx

	// server
	{
		serverId, err := SharedHTTPWebDAO.FindWebServerId(tx, 60)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("serverId:", serverId)
	}

	// location
	{
		serverId, err := SharedHTTPWebDAO.FindWebServerId(tx, 45)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("serverId:", serverId)
	}

	{
		serverId, err := SharedHTTPWebDAO.FindWebServerId(tx, 100)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("serverId:", serverId)
	}
}

func TestHTTPWebDAO_FindEnabledWebIdWithLocationId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithLocationId(tx, 17)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("webId:", webId)
}

func TestHTTPWebDAO_FindEnabledWebIdWithRewriteRuleId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithRewriteRuleId(tx, 13)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("webId:", webId)
}

func TestHTTPWebDAO_FindEnabledWebIdWithPageId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithPageId(tx, 15)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("webId:", webId)
}

func TestHTTPWebDAO_FindEnabledWebIdWithHeaderPolicyId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithHeaderPolicyId(tx, 52)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("webId:", webId)
}

func TestHTTPWebDAO_FindEnabledWebIdWithGzip(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithGzipId(tx, 9)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("webId:", webId)
}

func TestHTTPWebDAO_FindEnabledWebIdWithWebsocket(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	webId, err := SharedHTTPWebDAO.FindEnabledWebIdWithWebsocketId(tx, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("webId:", webId)
}

