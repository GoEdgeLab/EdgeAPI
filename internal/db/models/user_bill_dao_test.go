package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestUserBillDAO_GenerateBills(t *testing.T) {
	dbs.NotifyReady()

	err := SharedUserBillDAO.GenerateBills(timeutil.Format("Ym"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
