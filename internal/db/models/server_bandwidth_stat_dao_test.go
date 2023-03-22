package models_test

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestServerBandwidthStatDAO_UpdateServerBandwidth(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	err := dao.UpdateServerBandwidth(tx, 1, 1, timeutil.Format("Ymd"), timeutil.FormatTime("Hi", time.Now().Unix()/300*300), 1024, 300, 0, 0, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestSeverBandwidthStatDAO_InsertManyStats(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	var count = 1 // 测试时将此值设为一个比较大的数字
	for i := 0; i < count; i++ {
		if i%10000 == 0 {
			t.Log(i)
		}
		var day = timeutil.Format("Ymd", time.Now().AddDate(0, 0, -rands.Int(0, 200)))
		var minute = fmt.Sprintf("%02d%02d", rands.Int(0, 23), rands.Int(0, 59))
		err := dao.UpdateServerBandwidth(tx, 1, int64(rands.Int(1, 10000)), day, minute, 1024, 300, 0, 0, 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Log("ok")
}

func TestServerBandwidthStatDAO_FindMonthlyPercentile(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	t.Log(dao.FindMonthlyPercentile(tx, 23, timeutil.Format("Ym"), 95, false))
	t.Log(dao.FindMonthlyPercentile(tx, 23, timeutil.Format("Ym"), 95, true))
}

func TestServerBandwidthStatDAO_FindAllServerStatsWithMonth(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	stats, err := dao.FindAllServerStatsWithMonth(tx, 23, timeutil.Format("Ym"), false)
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		t.Logf("%+v", stat)
	}
}

func TestServerBandwidthStatDAO_FindAllServerStatsWithDay(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	stats, err := dao.FindAllServerStatsWithDay(tx, 23, timeutil.Format("Ymd"), false)
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		t.Logf("%+v", stat)
	}
}

func TestServerBandwidthStatDAO_Clean(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	var before = time.Now()
	err := dao.Clean(tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok", time.Since(before).Seconds()*1000, "ms")
}

func TestServerBandwidthStatDAO_FindHourlyBandwidthStats(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	stats, err := dao.FindHourlyBandwidthStats(tx, 23, 24, false)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(stats, t)
}

func TestServerBandwidthStatDAO_FindDailyBandwidthStats(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	stats, err := dao.FindDailyBandwidthStats(tx, 23, 14, false)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(stats, t)
}

func TestServerBandwidthStatDAO_FindBandwidthStatsBetweenDays(t *testing.T) {
	var dao = models.NewServerBandwidthStatDAO()
	var tx *dbs.Tx
	stats, err := dao.FindBandwidthStatsBetweenDays(tx, 23, timeutil.Format("Ymd", time.Now().AddDate(0, 0, -2)), timeutil.Format("Ymd"), false)
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		t.Log(stat.Day, stat.TimeAt, "bytes:", stat.Bytes, "bits:", stat.Bits)
	}
}
