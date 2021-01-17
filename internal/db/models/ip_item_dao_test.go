package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestIPItemDAO_NotifyClustersUpdate(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedIPItemDAO.NotifyClustersUpdate(tx, 28, NodeTaskTypeIPItemChanged)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
