package stats

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestTrafficDailyStatDAO_IncreaseDayBytes(t *testing.T) {
	dbs.NotifyReady()

	var now = time.Now()
	err := SharedTrafficDailyStatDAO.IncreaseDailyStat(nil, timeutil.Format("Ymd"), 1, 1, 1, 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok", time.Since(now).Seconds()*1000, "ms")
}

func TestTrafficDailyStatDAO_IncreaseIPs(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	err := SharedTrafficDailyStatDAO.IncreaseIPs(tx, timeutil.Format("Ymd"), 123)
	if err != nil {
		t.Fatal(err)
	}
}
