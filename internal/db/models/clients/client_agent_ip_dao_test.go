package clients_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/clients"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestClientAgentIPDAO_CreateIP(t *testing.T) {
	var dao = clients.NewClientAgentIPDAO()
	err := dao.CreateIP(nil, 1, "127.0.0.1", "")
	if err != nil {
		t.Fatal(err)
	}
}
