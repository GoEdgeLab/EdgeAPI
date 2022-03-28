package models_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/rands"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"testing"
	"time"
)

func TestNewMetricStatDAO_InsertMany(t *testing.T) {
	for i := 0; i <= 1; i++ {
		err := models.NewMetricStatDAO().CreateStat(nil, types.String(i)+"_v1", 18, int64(rands.Int(0, 10000)), int64(rands.Int(0, 10000)), int64(rands.Int(0, 100)), []string{"/html" + types.String(i)}, 1, timeutil.Format("Ymd"), 0)
		if err != nil {
			t.Fatal(err)
		}
		if i%10000 == 0 {
			t.Log(i)
		}
	}
	t.Log("done")
}

func TestMetricStatDAO_Clean2(t *testing.T) {
	dbs.NotifyReady()

	err := models.NewMetricStatDAO().Clean(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestMetricStatDAO_DeleteNodeItemStats(t *testing.T) {
	var dao = models.NewMetricStatDAO()
	var before = time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	err := dao.DeleteNodeItemStats(nil, 1, 0, 1, timeutil.Format("Ymd"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}

func TestMetricStatDAO_CountItemStats(t *testing.T) {
	var dao = models.NewMetricStatDAO()
	var before = time.Now()
	defer func() {
		t.Log(time.Since(before).Seconds()*1000, "ms")
	}()
	count, err := dao.CountItemStats(nil, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("count:", count)
}
