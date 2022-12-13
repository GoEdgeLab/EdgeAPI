package clients_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/clients"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestClientSystemDAO_CreateSystemIfNotExists(t *testing.T) {
	var dao = clients.NewClientSystemDAO()
	{
		err := dao.CreateSystemIfNotExists(nil, "Mac OS X")
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err := dao.CreateSystemIfNotExists(nil, "Mac OS X 2")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestClientSystemDAO_Clean(t *testing.T) {
	var dao = clients.NewClientSystemDAO()
	err := dao.Clean(nil, 30)
	if err != nil {
		t.Fatal(err)
	}
}
