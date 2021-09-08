package stats

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/assert"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestServerDomainHourlyStatDAO_PartitionTable(t *testing.T) {
	var a = assert.NewAssertion(t)

	var dao = NewServerDomainHourlyStatDAO()
	a.IsTrue(dao.PartitionTable("") == "edgeServerDomainHourlyStats_0")
	a.IsTrue(dao.PartitionTable("a1") == "edgeServerDomainHourlyStats_a")
	a.IsTrue(dao.PartitionTable("Y1") == "edgeServerDomainHourlyStats_y")
	a.IsTrue(dao.PartitionTable("z1") == "edgeServerDomainHourlyStats_z")
	a.IsTrue(dao.PartitionTable("A1") == "edgeServerDomainHourlyStats_a")
	a.IsTrue(dao.PartitionTable("Z1") == "edgeServerDomainHourlyStats_z")
	a.IsTrue(dao.PartitionTable("中国") == "edgeServerDomainHourlyStats_0")
	a.IsTrue(dao.PartitionTable("_") == "edgeServerDomainHourlyStats_0")
	a.IsTrue(dao.PartitionTable(" ") == "edgeServerDomainHourlyStats_0")
}

func TestServerDomainHourlyStatDAO_FindAllPartitionTables(t *testing.T) {
	var dao = NewServerDomainHourlyStatDAO()
	t.Log(dao.FindAllPartitionTables())
}

func TestServerDomainHourlyStatDAO_IncreaseHourlyStat(t *testing.T) {
	dbs.NotifyReady()

	for i := 0; i < 1_000_000; i++ {
		var f = string([]rune{int32(rands.Int('0', '9'))})

		err := NewServerDomainHourlyStatDAO().IncreaseHourlyStat(nil, 18, 48, 23, f+"rand"+types.String(i%500_000)+".com", timeutil.Format("Ymd")+fmt.Sprintf("%02d", rands.Int(0, 23)), 1, 1, 1, 1, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if i%10000 == 0 {
			t.Log(i)
		}
	}
}

func TestServerDomainHourlyStatDAO_FindTopDomainStats(t *testing.T) {
	var dao = NewServerDomainHourlyStatDAO()
	var before = time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	stats, err := dao.FindTopDomainStats(nil, timeutil.Format("Ymd00"), timeutil.Format("Ymd23"), 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		t.Log(stat.Domain, stat.CountRequests)
	}
}

func TestServerDomainHourlyStatDAO_Clean(t *testing.T) {
	var dao = NewServerDomainHourlyStatDAO()
	err := dao.Clean(nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
