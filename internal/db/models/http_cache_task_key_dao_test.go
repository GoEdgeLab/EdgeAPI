package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
)

func TestHTTPCacheTaskKeyDAO_CreateKey(t *testing.T) {
	var dao = models.NewHTTPCacheTaskKeyDAO()
	var tx *dbs.Tx
	_, err := dao.CreateKey(tx, 1, "a", "purge", "key", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestHTTPCacheTaskKeyDAO_UpdateKeyStatus(t *testing.T) {
	dbs.NotifyReady()

	var dao = models.NewHTTPCacheTaskKeyDAO()
	var tx *dbs.Tx
	var errString = "" // "this is error"
	err := dao.UpdateKeyStatus(tx, 3, 1, errString, []byte(`{"1":true}`))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
