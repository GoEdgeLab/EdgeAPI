package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestServerDAO_ComposeServerConfig(t *testing.T) {
	config, err := SharedServerDAO.ComposeServerConfig(1)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config, t)
}
