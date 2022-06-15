package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"runtime"
	"testing"
)

func TestIPListDAO_IncreaseVersion(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx

	dao := NewIPListDAO()
	version, err := dao.IncreaseVersion(tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("version:", version)
}

func TestIPListDAO_CheckUserIPList(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx

	{
		err := NewIPListDAO().CheckUserIPList(tx, 1, 100)
		if err == ErrNotFound {
			t.Log("not found")
		} else {
			t.Log(err)
		}
	}

	{
		err := NewIPListDAO().CheckUserIPList(tx, 1, 85)
		if err == ErrNotFound {
			t.Log("not found")
		} else {
			t.Log(err)
		}
	}

	{
		err := NewIPListDAO().CheckUserIPList(tx, 1, 17)
		if err == ErrNotFound {
			t.Log("not found")
		} else {
			t.Log(err)
		}
	}
}

func BenchmarkIPListDAO_IncreaseVersion(b *testing.B) {
	runtime.GOMAXPROCS(1)

	dbs.NotifyReady()

	var tx *dbs.Tx

	dao := NewIPListDAO()
	for i := 0; i < b.N; i++ {
		_, _ = dao.IncreaseVersion(tx)
	}
}

