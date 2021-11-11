package accounts

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestUserAccountDAO_PayBills(t *testing.T) {
	dbs.NotifyReady()

	err := NewUserAccountDAO().PayBills(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
