package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestServerBillDAO_CreateOrUpdateServerBill(t *testing.T) {
	var dao = NewServerBillDAO()
	var tx *dbs.Tx
	var month = timeutil.Format("Y02")
	err := dao.CreateOrUpdateServerBill(tx, 1, 2, month, 4, 5, 6, 7, 95, 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
