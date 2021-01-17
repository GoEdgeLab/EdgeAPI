package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestHTTPFirewallRuleSetDAO_FindRuleSetIdWithRuleId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	before := time.Now()
	setId, err := SharedHTTPFirewallRuleSetDAO.FindEnabledRuleSetIdWithRuleId(tx, 20)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("setId:", setId)
	t.Log(time.Since(before).Seconds()*1000, "ms")
}
