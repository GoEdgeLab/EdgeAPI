package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestNodeClusterDAO_DisableNodeCluster(t *testing.T) {
	dbs.NotifyReady()

	err := SharedNodeClusterDAO.DisableNodeCluster(nil, 46)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
