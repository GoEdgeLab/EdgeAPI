package regions

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestRegionProvinceDAO_FindProvinceIdWithProvinceName(t *testing.T) {
	dbs.NotifyReady()

	for i := 0; i < 5; i++ {
		now := time.Now()
		provinceId, err := SharedRegionProvinceDAO.FindProvinceIdWithNameCacheable(nil, 1, "安徽省")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(provinceId, time.Since(now).Seconds()*1000, "ms")
	}

	t.Log("====")
	for i := 0; i < 5; i++ {
		now := time.Now()
		provinceId, err := SharedRegionProvinceDAO.FindProvinceIdWithNameCacheable(nil, 2, "安徽省")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(provinceId, time.Since(now).Seconds()*1000, "ms")
	}
}
