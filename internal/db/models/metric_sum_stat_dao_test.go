package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestMetricSumStatDAO_Clean(t *testing.T) {
	dbs.NotifyReady()

	err := NewMetricSumStatDAO().Clean(nil)
	if err != nil {
		t.Fatal(err)
	}
}
