package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestReverseProxyDAO_ComposeReverseProxyConfig(t *testing.T) {
	var tx *dbs.Tx
	config, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(tx, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
}
