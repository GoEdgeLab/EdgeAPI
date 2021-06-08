package models

import (
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestServerDailyStatDAO_SaveStats(t *testing.T) {
	var tx *dbs.Tx
	stats := []*pb.ServerDailyStat{
		{
			ServerId:  1,
			RegionId:  2,
			Bytes:     1,
			CreatedAt: 1607671488,
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
			ServerId:  1,
			RegionId:  3,
			Bytes:     1,
			CreatedAt: 1607671488,
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
	bytes, err := NewServerDailyStatDAO().SumUserMonthly(tx, 1, 1, timeutil.Format("Ym"))
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

	stat, err := NewServerDailyStatDAO().SumMinutelyStat(tx, 23, timeutil.Format("Ymd") + "1435")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(stat, t)
}
