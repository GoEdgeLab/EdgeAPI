package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHTTPWebDAO_UpdateWebShutdown(t *testing.T) {
	{
		err := SharedHTTPWebDAO.UpdateWebShutdown(1, []byte("{}"))
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		err := SharedHTTPWebDAO.UpdateWebShutdown(1, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Log("ok")
}

func TestHTTPWebDAO_FindAllWebIdsWithHTTPFirewallPolicyId(t *testing.T) {
	dbs.NotifyReady()

	webIds, err := SharedHTTPWebDAO.FindAllWebIdsWithHTTPFirewallPolicyId(9)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("webIds:", webIds)

	count, err := SharedServerDAO.CountEnabledServersWithWebIds(webIds)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("count:", count)
}
