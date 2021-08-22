package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestReverseProxyDAO_ComposeReverseProxyConfig(t *testing.T) {
	var tx *dbs.Tx
	config, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
}

func TestReverseProxyDAO_FindReverseProxyContainsOriginId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	reverseProxyId, err := SharedReverseProxyDAO.FindReverseProxyContainsOriginId(tx, 68)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("reverseProxyId:", reverseProxyId)
}
