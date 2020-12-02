package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestLogDAO_SumLogsSize(t *testing.T) {
	dbs.NotifyReady()

	size, err := SharedLogDAO.SumLogsSize()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("size:", size)
}
