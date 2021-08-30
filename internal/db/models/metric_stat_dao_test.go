package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	"testing"
)

func TestNewMetricStatDAO_InsertMany(t *testing.T) {
	for i := 0; i <= 10_000_000; i++ {
		err := NewMetricStatDAO().CreateStat(nil, types.String(i)+"_v1", 18, int64(rands.Int(0, 10000)), int64(rands.Int(0, 10000)), int64(rands.Int(0, 100)), []string{"/html" + types.String(i)}, 1, "20210830", 0)
		if err != nil {
			t.Fatal(err)
		}
		if i % 10000 == 0 {
			t.Log(i)
		}
	}
	t.Log("done")
}
