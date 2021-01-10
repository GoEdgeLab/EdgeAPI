package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"runtime"
	"testing"
)

func TestAPINodeDAO_FindEnabledAPINodeIdWithAddr(t *testing.T) {
	dao := NewAPINodeDAO()
	var tx *dbs.Tx
	{
		apiNodeId, err := dao.FindEnabledAPINodeIdWithAddr(tx, "http", "127.0.0.1", 123)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("apiNodeId:", apiNodeId)
	}

	{
		apiNodeId, err := dao.FindEnabledAPINodeIdWithAddr(tx, "http", "127.0.0.1", 8003)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("apiNodeId:", apiNodeId)
	}
}

func BenchmarkAPINodeDAO_New(b *testing.B) {
	runtime.GOMAXPROCS(1)
	for i := 0; i < b.N; i++ {
		_ = NewAPINodeDAO()
	}
}
