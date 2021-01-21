package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestTrafficDailyStatDAO_IncreaseDayBytes(t *testing.T) {
	dbs.NotifyReady()

	now := time.Now()
	err := SharedTrafficDailyStatDAO.IncreaseDayBytes(nil, timeutil.Format("Ymd"), 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok", time.Since(now).Seconds()*1000, "ms")
}
