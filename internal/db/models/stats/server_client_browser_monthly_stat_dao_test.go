package stats

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestServerClientBrowserMonthlyStatDAO_IncreaseMonthlyCount(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedServerClientBrowserMonthlyStatDAO.IncreaseMonthlyCount(tx, 1, 1, "1.0", "202101", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
