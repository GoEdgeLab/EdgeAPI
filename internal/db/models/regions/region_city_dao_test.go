package regions

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"testing"
	"time"
)

func TestRegionCityDAO_FindCityIdWithCityNameCacheable(t *testing.T) {
	dbs.NotifyReady()

	for i := 0; i < 5; i++ {
		now := time.Now()
		cityId, err := SharedRegionCityDAO.FindCityIdWithNameCacheable(nil, 1, "北京市")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("cityId", cityId, time.Since(now).Seconds()*1000, "ms")
	}
}
