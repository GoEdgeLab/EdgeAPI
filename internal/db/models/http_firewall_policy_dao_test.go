package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHTTPFirewallPolicyDAO_FindFirewallPolicyIdsContainsIPList(t *testing.T) {
	dbs.NotifyReady()

	{
		policyIds, err := SharedHTTPFirewallPolicyDAO.FindEnabledFirewallPolicyIdsWithIPListId(nil, 8)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(policyIds)
	}

	{
		policyIds, err := SharedHTTPFirewallPolicyDAO.FindEnabledFirewallPolicyIdsWithIPListId(nil, 18)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(policyIds)
	}
}

func TestHTTPFirewallPolicyDAO_FindEnabledFirewallPolicyIdWithRuleGroupId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	policyIds, err := SharedHTTPFirewallPolicyDAO.FindEnabledFirewallPolicyIdWithRuleGroupId(tx, 160)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("policyIds:", policyIds)
}
