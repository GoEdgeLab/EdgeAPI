package models

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestSysLockerDAO_Lock(t *testing.T) {
	isOk, err := SharedSysLockerDAO.Lock("test", 600)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(isOk)

	if isOk {
		err = SharedSysLockerDAO.Unlock("test")
		if err != nil {
			t.Fatal(err)
		}
	}
}
