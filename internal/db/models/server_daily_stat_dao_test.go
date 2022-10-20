package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestServerDailyStatDAO_SaveStats(t *testing.T) {
	var tx *dbs.Tx
	stats := []*pb.ServerDailyStat{
		{
			ServerId:     1,
			NodeRegionId: 2,
			Bytes:        1,
			CreatedAt:    1607671488,
		},
	}
	err := NewServerDailyStatDAO().SaveStats(tx, stats)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestServerDailyStatDAO_SaveStats2(t *testing.T) {
	var tx *dbs.Tx
	stats := []*pb.ServerDailyStat{
		{
			ServerId:     1,
			NodeRegionId: 3,
			Bytes:        1,
			CreatedAt:    1607671488,
		},
	}
	err := NewServerDailyStatDAO().SaveStats(tx, stats)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestServerDailyStatDAO_SumUserMonthly(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx
	bytes, err := NewServerDailyStatDAO().SumUserMonthly(tx, 1, timeutil.Format("Ym"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("bytes:", bytes)
}

func TestServerDailyStatDAO_SumHourlyRequests(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx

	stat, err := NewServerDailyStatDAO().SumHourlyStat(tx, 23, timeutil.Format("YmdH"))
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(stat, t)
}

func TestServerDailyStatDAO_SumMinutelyRequests(t *testing.T) {
	dbs.NotifyReady()
	var tx *dbs.Tx

	stat, err := NewServerDailyStatDAO().SumMinutelyStat(tx, 23, timeutil.Format("Ymd")+"1435")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(stat, t)
}

func TestServerDailyStatDAO_FindDistinctPlanServerIdsBetweenDay(t *testing.T) {
	var tx *dbs.Tx
	serverIds, err := NewServerDailyStatDAO().FindDistinctServerIds(tx, timeutil.Format("Ym01"), timeutil.Format("Ymd"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(serverIds)
}

func TestServerDailyStatDAO_FindStatsBetweenDays(t *testing.T) {
	var tx *dbs.Tx
	stats, err := NewServerDailyStatDAO().FindStatsBetweenDays(tx, 1, 0, 0, timeutil.Format("Ymd", time.Now().AddDate(0, 0, -1)), timeutil.Format("Ymd"))
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		t.Log(stat.Day, stat.TimeFrom, stat.TimeTo, stat.Bytes)
	}
}

func TestServerDailyStatDAO_FindStatsWithDay(t *testing.T) {
	var dao = NewServerDailyStatDAO()
	var tx *dbs.Tx
	stats, err := dao.FindStatsWithDay(tx, 23, timeutil.Format("Ymd"), "000000", "235900")
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		t.Log(stat.TimeFrom, stat.TimeTo, stat.Bytes)
	}
}
