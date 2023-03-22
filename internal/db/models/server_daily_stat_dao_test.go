package models_test

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestServerDailyStatDAO_SaveStats(t *testing.T) {
	var tx *dbs.Tx
	var stats = []*pb.ServerDailyStat{
		{
			ServerId:     1,
			NodeRegionId: 2,
			Bytes:        1,
			CreatedAt:    1607671488,
		},
	}
	err := models.NewServerDailyStatDAO().SaveStats(tx, stats)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestServerDailyStatDAO_SaveStats2(t *testing.T) {
	var tx *dbs.Tx
	var stats = []*pb.ServerDailyStat{
		{
			ServerId:     1,
			NodeRegionId: 3,
			Bytes:        1,
			CreatedAt:    1607671488,
		},
	}
	err := models.NewServerDailyStatDAO().SaveStats(tx, stats)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestServerDailyStatDAO_SumUserMonthly(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	bytes, err := models.NewUserBandwidthStatDAO().SumUserMonthly(tx, 1, timeutil.Format("Ym"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("bytes:", bytes)
}

func TestServerDailyStatDAO_SumHourlyRequests(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx

	stat, err := models.NewServerDailyStatDAO().SumHourlyStat(tx, 23, timeutil.Format("YmdH"))
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(stat, t)
}

func TestServerDailyStatDAO_SumMinutelyRequests(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx

	stat, err := models.NewServerDailyStatDAO().SumMinutelyStat(tx, 23, timeutil.Format("Ymd")+"1435")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(stat, t)
}

func TestServerDailyStatDAO_FindDistinctPlanServerIdsBetweenDay(t *testing.T) {
	var tx *dbs.Tx
	serverIds, err := models.NewServerDailyStatDAO().FindDistinctServerIds(tx, timeutil.Format("Ym01"), timeutil.Format("Ymd"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(serverIds)
}

func TestServerDailyStatDAO_FindStatsBetweenDays(t *testing.T) {
	var tx *dbs.Tx
	stats, err := models.NewServerDailyStatDAO().FindStatsBetweenDays(tx, 1, 0, 0, timeutil.Format("Ymd", time.Now().AddDate(0, 0, -1)), timeutil.Format("Ymd"))
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		t.Log(stat.Day, stat.TimeFrom, stat.TimeTo, stat.Bytes)
	}
}

func TestSeverDailyStatDAO_InsertMany(t *testing.T) {
	dbs.NotifyReady()

	var tx *dbs.Tx
	var dao = models.NewServerDailyStatDAO()
	var count = 1 // 实际测试时可以将此值调的很大
	for i := 0; i < count; i++ {
		if i%10000 == 0 {
			t.Log(i)
		}
		err := dao.SaveStats(tx, []*pb.ServerDailyStat{{
			ServerId:             23,
			NodeRegionId:         int64(rands.Int(0, 999999)),
			Bytes:                1024,
			CachedBytes:          1024,
			CountRequests:        1024,
			CountCachedRequests:  1024,
			CreatedAt:            time.Now().Unix(),
			CountAttackRequests:  1024,
			AttackBytes:          1024,
			CheckTrafficLimiting: false,
			PlanId:               0,
			Day:                  "202303" + fmt.Sprintf("%02d", rands.Int(1, 31)),
			Hour:                 "2023032101",
			TimeFrom:             fmt.Sprintf("%06d", rands.Int(0, 999999)),
			TimeTo:               "211459",
		}})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestServerDailyStatDAO_FindStatsWithDay(t *testing.T) {
	var dao = models.NewServerDailyStatDAO()
	var tx *dbs.Tx
	stats, err := dao.FindStatsWithDay(tx, 23, timeutil.Format("Ymd"), "000000", "235900")
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		t.Log(stat.TimeFrom, stat.TimeTo, stat.Bytes)
	}
}
