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