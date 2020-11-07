package models

import (
	_ "github.com/go-sql-driver/mysql"
	"runtime"
	"testing"
)

func TestIPListDAO_IncreaseVersion(t *testing.T) {
	dao := NewIPListDAO()
	version, err := dao.IncreaseVersion(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("version:", version)
}

func BenchmarkIPListDAO_IncreaseVersion(b *testing.B) {
	runtime.GOMAXPROCS(1)

	dao := NewIPListDAO()
	for i := 0; i < b.N; i++ {
		_, _ = dao.IncreaseVersion(1)
	}
}
