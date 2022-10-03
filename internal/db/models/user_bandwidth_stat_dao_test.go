package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
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

func TestUserBandwidthStatDAO_FindBandwidthStatsBetweenDays(t *testing.T) {
	var dao = models.NewUserBandwidthStatDAO()
	var tx *dbs.Tx
	stats, err := dao.FindUserBandwidthStatsBetweenDays(tx, 1, timeutil.Format("Ymd", time.Now().AddDate(0, 0, -2)), timeutil.Format("Ymd"))
	if err != nil {
		t.Fatal(err)
	}
	var dataCount = 0
	for _, stat := range stats {
		t.Log(stat.Day, stat.TimeAt, "bytes:", stat.Bytes, "bits:", stat.Bits)
		if stat.Bytes > 0 {
			dataCount++
		}
	}
	t.Log("data count:", dataCount)
}
