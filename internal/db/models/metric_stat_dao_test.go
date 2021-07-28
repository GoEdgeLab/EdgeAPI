package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/types"
	"testing"
)

func TestNewMetricStatDAO_InsertMany(t *testing.T) {
	for i := 0; i <= 1; i++ {
		err := NewMetricStatDAO().CreateStat(nil, types.String(i) + "_v1", 18, 48, 23, 25, []string{"/html" + types.String(i)}, 1, "20210728", 0)
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Log("done")
}
