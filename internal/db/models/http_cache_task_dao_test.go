package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHTTPCacheTaskDAO_Clean(t *testing.T) {
	dbs.NotifyReady()

	err := models.SharedHTTPCacheTaskDAO.Clean(nil, 30)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
