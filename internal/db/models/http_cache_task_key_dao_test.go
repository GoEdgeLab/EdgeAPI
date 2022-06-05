package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
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

func TestHTTPCacheTaskKeyDAO_CountUserTasksInDay(t *testing.T) {
	dbs.NotifyReady()

	var dao = models.NewHTTPCacheTaskKeyDAO()
	var tx *dbs.Tx
	{
		count, err := dao.CountUserTasksInDay(tx, 1, timeutil.Format("Ymd"), models.HTTPCacheTaskTypePurge)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("count:", count)
	}
	{
		count, err := dao.CountUserTasksInDay(tx, 1, timeutil.Format("Ymd"), models.HTTPCacheTaskTypeFetch)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("count:", count)
	}
}
