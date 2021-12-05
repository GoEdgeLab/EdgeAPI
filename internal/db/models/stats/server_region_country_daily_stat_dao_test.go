package stats

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
)

func TestServerRegionCountryDailyStatDAO_IncreaseDailyStat(t *testing.T) {
	var tx *dbs.Tx
	err := NewServerRegionCountryDailyStatDAO().IncreaseDailyStat(tx, 1, 3, timeutil.Format("Ymd"), 2, 2, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestServerRegionCountryDailyStatDAO_ListSumStats(t *testing.T) {
	var tx *dbs.Tx
	stats, err := NewServerRegionCountryDailyStatDAO().ListSumStats(tx, timeutil.Format("Ymd"), "countAttackRequests", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, stat := range stats {
		statJSON, err := json.Marshal(stat)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(statJSON))
	}
}
