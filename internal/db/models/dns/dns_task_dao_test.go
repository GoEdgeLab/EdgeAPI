package dns

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestDNSTaskDAO_CreateDNSTask(t *testing.T) {
	dbs.NotifyReady()
	err := SharedDNSTaskDAO.CreateDNSTask(nil, 1, 2, 3, 0, "taskType")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
