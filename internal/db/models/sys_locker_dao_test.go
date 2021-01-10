package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestSysLockerDAO_Lock(t *testing.T) {
	var tx *dbs.Tx

	isOk, err := SharedSysLockerDAO.Lock(tx, "test", 600)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(isOk)

	if isOk {
		err = SharedSysLockerDAO.Unlock(tx, "test")
		if err != nil {
			t.Fatal(err)
		}
	}
}
