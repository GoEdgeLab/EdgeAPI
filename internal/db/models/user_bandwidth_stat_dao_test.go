package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestUserBandwidthStatDAO_UpdateServerBandwidth(t *testing.T) {
	var dao = models.NewUserBandwidthStatDAO()
	var tx *dbs.Tx
	err := dao.UpdateUserBandwidth(tx, 1, timeutil.Format("Ymd"), timeutil.Format("Hi"), 1024)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestUserBandwidthStatDAO_Clean(t *testing.T) {
	var dao = models.NewUserBandwidthStatDAO()
	var tx *dbs.Tx
	err := dao.Clean(tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
