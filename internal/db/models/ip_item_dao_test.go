package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestIPItemDAO_NotifyClustersUpdate(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedIPItemDAO.NotifyUpdate(tx, 28)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestIPItemDAO_DisableIPItemsWithListId(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedIPItemDAO.DisableIPItemsWithListId(tx, 67)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
