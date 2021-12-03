package models

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestMetricSumStatDAO_Clean(t *testing.T) {
	err := NewMetricSumStatDAO().Clean(nil, 20)
	if err != nil {
		t.Fatal(err)
	}
}
