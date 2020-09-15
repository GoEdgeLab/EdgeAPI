package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestReverseProxyDAO_ComposeReverseProxyConfig(t *testing.T) {
	config, err := SharedReverseProxyDAO.ComposeReverseProxyConfig(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
}
